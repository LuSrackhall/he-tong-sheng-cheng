package resolver

import (
	"fmt"

	"asset-leasing-system/runtime/internal/model"
)

// Resolver 负责引用解析（BR → Rule、workflow → capability DAG）
type Resolver struct {
	capabilities map[string]*model.Capability
	rules        map[string]*model.Rule
	workflows    map[string]*model.Workflow
}

func New(caps []*model.Capability, rules []*model.Rule, wfs []*model.Workflow) *Resolver {
	capMap := make(map[string]*model.Capability, len(caps))
	for _, c := range caps {
		capMap[c.ID] = c
	}
	ruleMap := make(map[string]*model.Rule, len(rules))
	for _, r := range rules {
		ruleMap[r.ID] = r
	}
	wfMap := make(map[string]*model.Workflow, len(wfs))
	for _, w := range wfs {
		wfMap[w.ID] = w
	}
	return &Resolver{
		capabilities: capMap,
		rules:        ruleMap,
		workflows:    wfMap,
	}
}

// ResolveCapability 解析 Capability 的所有引用
func (r *Resolver) ResolveCapability(id string) (*model.ResolvedCapability, error) {
	cap, ok := r.capabilities[id]
	if !ok {
		return nil, fmt.Errorf("capability %q not found", id)
	}

	rc := &model.ResolvedCapability{
		Capability: cap,
		DAG:        make(map[string][]string),
	}

	// 解析 Rule 引用
	for _, brID := range cap.Preconditions {
		rule, err := r.resolveRule(brID)
		if err != nil {
			return nil, fmt.Errorf("capability %q: %w", id, err)
		}
		rc.ResolvedRules = append(rc.ResolvedRules, rule)
	}

	// 构建依赖 DAG
	rc.Dependencies = cap.Dependencies.Requires
	for _, depID := range cap.Dependencies.Requires {
		if _, ok := r.capabilities[depID]; !ok {
			return nil, fmt.Errorf("capability %q depends on %q which is not defined", id, depID)
		}
		rc.DAG[depID] = cap.Dependencies.Triggers
	}

	return rc, nil
}

// ResolveWorkflow 展开 Workflow 为 Capability 列表
func (r *Resolver) ResolveWorkflow(id string) ([]string, error) {
	wf, ok := r.workflows[id]
	if !ok {
		return nil, fmt.Errorf("workflow %q not found", id)
	}

	var steps []string
	for _, step := range wf.Steps {
		if step.Capability != "" {
			if _, ok := r.capabilities[step.Capability]; !ok {
				return nil, fmt.Errorf("workflow %q references unknown capability %q", id, step.Capability)
			}
			steps = append(steps, step.Capability)

			// 展开依赖
			depSteps := r.expandDependencies(step.Capability)
			steps = append(steps, depSteps...)
		}
	}

	return steps, nil
}

// expandDependencies 展开 Capability 的依赖链
func (r *Resolver) expandDependencies(capID string) []string {
	var deps []string
	cap, ok := r.capabilities[capID]
	if !ok {
		return deps
	}
	for _, depID := range cap.Dependencies.Requires {
		if _, ok := r.capabilities[depID]; ok {
			deps = append(deps, depID)
		}
	}
	return deps
}

// ValidateIntegrity 校验所有引用的完整性
func (r *Resolver) ValidateIntegrity() []error {
	var errs []error

	// 验证 Capability 引用的规则
	for _, cap := range r.capabilities {
		for _, brID := range cap.Preconditions {
			if err := r.validateRuleRef(cap.ID, brID); err != nil {
				errs = append(errs, err)
			}
		}
		for _, depID := range cap.Dependencies.Requires {
			if _, ok := r.capabilities[depID]; !ok {
				errs = append(errs, fmt.Errorf("capability %q: dependency %q not found", cap.ID, depID))
			}
		}
	}

	// 验证 Workflow 引用的 Capability
	for _, wf := range r.workflows {
		for _, step := range wf.Steps {
			if step.Capability != "" {
				if _, ok := r.capabilities[step.Capability]; !ok {
					errs = append(errs, fmt.Errorf("workflow %q: capability %q not found", wf.ID, step.Capability))
				}
			}
		}
	}

	// 验证 Rule 的 applies_to
	for _, rule := range r.rules {
		for _, capID := range rule.AppliesTo {
			if _, ok := r.capabilities[capID]; !ok {
				errs = append(errs, fmt.Errorf("rule %q references unknown capability %q", rule.ID, capID))
			}
		}
	}

	return errs
}

func (r *Resolver) resolveRule(id string) (*model.Rule, error) {
	rule, ok := r.rules[id]
	if !ok {
		return nil, fmt.Errorf("rule %q not found", id)
	}
	return rule, nil
}

func (r *Resolver) validateRuleRef(capID, ruleID string) error {
	if _, ok := r.rules[ruleID]; !ok {
		return fmt.Errorf("capability %q: precondition %q not found", capID, ruleID)
	}
	return nil
}

// DetectCycle 检测 Capability 依赖 DAG 中是否存在环
func (r *Resolver) DetectCycle() []error {
	var errs []error
	for id := range r.capabilities {
		visited := make(map[string]bool)
		path := make(map[string]bool)
		if r.hasCycle(id, visited, path) {
			errs = append(errs, fmt.Errorf("cycle detected in dependency graph involving %q", id))
		}
	}
	return errs
}

func (r *Resolver) hasCycle(id string, visited, path map[string]bool) bool {
	if path[id] {
		return true
	}
	if visited[id] {
		return false
	}
	visited[id] = true
	path[id] = true

	if cap, ok := r.capabilities[id]; ok {
		for _, depID := range cap.Dependencies.Requires {
			if r.hasCycle(depID, visited, path) {
				return true
			}
		}
	}

	path[id] = false
	return false
}
