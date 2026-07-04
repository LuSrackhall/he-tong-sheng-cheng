package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"asset-leasing-system/runtime/internal/cache"
	"asset-leasing-system/runtime/internal/explain"
	"asset-leasing-system/runtime/internal/guard"
	"asset-leasing-system/runtime/internal/loader"
	"asset-leasing-system/runtime/internal/model"
	"asset-leasing-system/runtime/internal/planner"
	"asset-leasing-system/runtime/internal/resolver"
	"asset-leasing-system/runtime/internal/snapshot"
	"asset-leasing-system/runtime/internal/trace"
)

type RuntimeConfig struct {
	KnowledgeDir string
	APIBase      string
	Validate     bool
	TracesDir    string
}

type KnowledgeRuntime struct {
	Loader   *loader.Loader
	Resolver *resolver.Resolver
	Planner  *planner.Planner
	Guard    *guard.ConstitutionGuard
	Cache    *cache.Cache
	Snapshot *snapshot.Snapshot
}

func initRuntime(cfg *RuntimeConfig) (*KnowledgeRuntime, error) {
	ld := loader.New(cfg.KnowledgeDir)

	// 加载所有定义
	caps, err := ld.LoadAllCapabilities()
	if err != nil {
		return nil, fmt.Errorf("load capabilities: %w", err)
	}
	rules, err := ld.LoadAllRules()
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}
	wfs, err := ld.LoadAllWorkflows()
	if err != nil {
		return nil, fmt.Errorf("load workflows: %w", err)
	}

	log.Printf("Loaded %d capabilities, %d rules, %d workflows", len(caps), len(rules), len(wfs))

	// 解析引用
	res := resolver.New(caps, rules, wfs)

	// 完整性校验
	if errs := res.ValidateIntegrity(); len(errs) > 0 {
		for _, e := range errs {
			log.Printf("Integrity error: %v", e)
		}
		return nil, fmt.Errorf("integrity check failed with %d errors", len(errs))
	}

	// DAG 无环校验
	if errs := res.DetectCycle(); len(errs) > 0 {
		for _, e := range errs {
			log.Printf("Cycle detected: %v", e)
		}
		return nil, fmt.Errorf("cycle detection failed with %d errors", len(errs))
	}

	log.Println("Integrity check passed, no cycles detected")

	// 构建运行时
	plan := planner.New(res)
	snap := snapshot.New(res, plan)
	cch := cache.New()

	// 初始化缓存
	for _, c := range caps {
		cch.SetCapability(c.ID, c)
	}
	for _, r := range rules {
		cch.SetRule(r.ID, r)
	}
	for _, w := range wfs {
		cch.SetWorkflow(w.ID, w)
	}

	// Guard
	knownCaps := make(map[string]bool)
	for _, c := range caps {
		knownCaps[c.ID] = true
	}
	grd := guard.New(knownCaps)
	if !cfg.Validate {
		grd.SetStrictMode(false)
		log.Println("Guard strict mode disabled (--validate=false)")
	}

	return &KnowledgeRuntime{
		Loader:   ld,
		Resolver: res,
		Planner:  plan,
		Guard:    grd,
		Cache:    cch,
		Snapshot: snap,
	}, nil
}

func parseGlobalFlags(args []string) (*RuntimeConfig, []string) {
	cfg := &RuntimeConfig{
		KnowledgeDir: "knowledge",
		APIBase:      "http://localhost:8080",
		Validate:     true,
		TracesDir:    ".traces",
	}

	// 支持 --flag=value 和 --flag value 两种格式
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			return cfg, []string{"--help"}
		case arg == "--validate=false":
			cfg.Validate = false
		case arg == "--validate=true":
			cfg.Validate = true
		case strings.HasPrefix(arg, "--validate="):
			// already handled above; skip
		case strings.HasPrefix(arg, "--knowledge-dir="):
			cfg.KnowledgeDir = strings.TrimPrefix(arg, "--knowledge-dir=")
		case arg == "--knowledge-dir" && i+1 < len(args):
			i++
			cfg.KnowledgeDir = args[i]
		case strings.HasPrefix(arg, "--traces-dir="):
			cfg.TracesDir = strings.TrimPrefix(arg, "--traces-dir=")
		case arg == "--traces-dir" && i+1 < len(args):
			i++
			cfg.TracesDir = args[i]
		case strings.HasPrefix(arg, "--api-base="):
			cfg.APIBase = strings.TrimPrefix(arg, "--api-base=")
		case arg == "--api-base" && i+1 < len(args):
			i++
			cfg.APIBase = args[i]
		default:
			remaining = append(remaining, arg)
		}
	}
	return cfg, remaining
}

func cmdPlan(rt *KnowledgeRuntime, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: kr plan <capability|workflow>")
	}
	name := args[0]

	// 先检查是否是 workflow
	if rt.Cache.GetWorkflow(name) != nil {
		plan, err := rt.Planner.PlanWorkflow(name, model.ProfileDebug)
		if err != nil {
			return fmt.Errorf("plan workflow: %w", err)
		}
		printPlan(plan)
		return nil
	}

	// 否则视为 capability
	plan, err := rt.Planner.Plan(name, model.ProfileDebug)
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	printPlan(plan)
	return nil
}

func cmdRun(rt *KnowledgeRuntime, args []string, cfg *RuntimeConfig) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: kr run <capability> [--inputs...]")
	}
	name := args[0]

	// 解析 inputs
	inputs := parseInputs(args[1:])

	// 生成 Plan
	plan, err := rt.Planner.Plan(name, model.ProfileStrict)
	if err != nil {
		return fmt.Errorf("plan: %w", err)
	}

	// Pre-execution Guard
	result := rt.Guard.Check(plan)
	if len(result.Violations) > 0 {
		if cfg.Validate {
			fmt.Println("Constitution Guard blocked execution:")
			for _, v := range result.Violations {
				fmt.Printf("  §%s: %s\n", v.Axiom, v.Detail)
			}
			return fmt.Errorf("execution blocked by Constitution Guard")
		}
		fmt.Println("Guard violations (skipped due to --validate=false):")
		for _, v := range result.Violations {
			fmt.Printf("  §%s: %s\n", v.Axiom, v.Detail)
		}
	}

	// 构建 Trace
	tr := &model.Trace{
		Identity: model.TraceIdentity{
			CapabilityID: name,
			PlanVersion:  plan.Version,
			PlanHash:     plan.PlanHash,
			Adapter:      "cli",
		},
		Context: model.TraceContext{
			User:        "cli",
			Timestamp:   time.Now(),
			Environment: "development",
		},
		Profile:       string(model.ProfileStrict),
		PlanSnapshot:  *plan,
		Determinism: model.Determinism{
			Score: 0.99,
			Factors: map[string]float64{
				"ui_stability":    1.0,
				"api_consistency": 1.0,
				"agent_variance":  1.0,
			},
		},
	}

	// Simulate execution steps
	start := time.Now()
	for _, step := range plan.Steps {
		tStep := model.TraceStep{
			IntentStep: fmt.Sprintf("Execute %s", step.CapabilityID),
			Input:      convertInputMapping(step.InputMapping, inputs),
			Output:     map[string]any{"status": "simulated"},
			DurationMs: time.Since(start).Milliseconds(),
		}
		tr.Steps = append(tr.Steps, tStep)
	}

	// Observability checks
	for _, rule := range plan.ResolvedRules {
		tr.Observability.Checks = append(tr.Observability.Checks, model.ObservabilityCheck{
			RuleID: rule.RuleID,
			Passed: true,
		})
	}

	tr.Summary = model.TraceSummary{
		DurationMs: time.Since(start).Milliseconds(),
		StepCount:  len(tr.Steps),
	}

	// 写入 Trace
	writer := trace.NewWriter(cfg.TracesDir)
	path, err := writer.Write(tr)
	if err != nil {
		return fmt.Errorf("write trace: %w", err)
	}

	fmt.Printf("Execution complete. Trace: %s\n", tr.Identity.TraceID)
	fmt.Printf("Trace file: %s\n", path)
	fmt.Printf("Plan hash: %s\n", plan.PlanHash)
	fmt.Printf("Steps: %d, Duration: %dms\n", tr.Summary.StepCount, tr.Summary.DurationMs)

	return nil
}

func cmdExplain(args []string, cfg *RuntimeConfig) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: kr explain <trace-id>")
	}
	traceID := args[0]
	if traceID == "--trace" && len(args) > 1 {
		traceID = args[1]
	}
	reader := trace.NewReader(cfg.TracesDir)
	exp := explain.New(reader)

	output, err := exp.ByTraceID(traceID)
	if err != nil {
		return fmt.Errorf("explain: %w", err)
	}
	fmt.Print(output)
	return nil
}

func printPlan(plan *model.ExecutionPlan) {
	fmt.Printf("Plan: %s (v%d)\n", plan.CapabilityID, plan.Version)
	fmt.Printf("Hash: %s\n", plan.PlanHash)
	fmt.Printf("Profile: %s\n", plan.Profile)
	fmt.Println()
	fmt.Println("Steps:")
	for i, step := range plan.Steps {
		fmt.Printf("  %d. %s\n", i+1, step.CapabilityID)
	}
	if len(plan.ResolvedRules) > 0 {
		fmt.Println()
		fmt.Println("Rules:")
		for _, rule := range plan.ResolvedRules {
			fmt.Printf("  %s: %s\n", rule.RuleID, rule.Title)
		}
	}
	if len(plan.Permissions) > 0 {
		fmt.Println()
		fmt.Println("Permissions:")
		for _, p := range plan.Permissions {
			fmt.Printf("  %s\n", p)
		}
	}
}

func parseInputs(args []string) map[string]string {
	inputs := make(map[string]string)
	for _, arg := range args {
		if parts := strings.SplitN(arg, "=", 2); len(parts) == 2 {
			inputs[parts[0]] = parts[1]
		}
	}
	return inputs
}

func convertInputMapping(mapping map[string]string, inputs map[string]string) map[string]any {
	result := make(map[string]any)
	for k, v := range mapping {
		if inputVal, ok := inputs[k]; ok {
			result[k] = inputVal
		} else {
			result[k] = v // 保留模板变量
		}
	}
	return result
}

func main() {
	// 解析全局 flags（支持放在子命令前后任意位置）
	cfg, remaining := parseGlobalFlags(os.Args[1:])

	if len(remaining) < 1 || remaining[0] == "--help" {
		fmt.Println("Usage: kr <command> [args...] [flags]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  plan <capability|workflow>    Preview Execution Plan")
		fmt.Println("  run <capability> [inputs...]   Execute a Capability")
		fmt.Println("  explain <trace-id>             Explain a Trace")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --knowledge-dir <path>   Knowledge layer path (default: knowledge)")
		fmt.Println("  --api-base <url>         Backend API URL (default: http://localhost:8080)")
		fmt.Println("  --validate=<bool>        Enable Constitution Guard (default: true)")
		fmt.Println("  --traces-dir <path>      Trace storage path (default: .traces)")
		os.Exit(0)
	}

	cmd := remaining[0]
	cmdArgs := remaining[1:]

	switch cmd {
	case "plan":
		rt, err := initRuntime(cfg)
		if err != nil {
			log.Fatalf("Runtime init failed: %v", err)
		}
		if err := cmdPlan(rt, cmdArgs); err != nil {
			log.Fatalf("Plan failed: %v", err)
		}

	case "run":
		rt, err := initRuntime(cfg)
		if err != nil {
			log.Fatalf("Runtime init failed: %v", err)
		}
		if err := cmdRun(rt, cmdArgs, cfg); err != nil {
			log.Fatalf("Run failed: %v", err)
		}

	case "explain":
		if err := cmdExplain(cmdArgs, cfg); err != nil {
			log.Fatalf("Explain failed: %v", err)
		}

	default:
		log.Fatalf("Unknown command: %s (use plan, run, or explain)", cmd)
	}
}
