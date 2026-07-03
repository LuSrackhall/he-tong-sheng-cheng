package planner

import (
	"crypto/sha256"
	"fmt"
	"time"

	"asset-leasing-system/runtime/internal/model"
	"asset-leasing-system/runtime/internal/resolver"
)

// Planner 负责将 Capability 展开为 Execution Plan
type Planner struct {
	resolver *resolver.Resolver
}

func New(res *resolver.Resolver) *Planner {
	return &Planner{resolver: res}
}

// Plan 为单个 Capability 生成 Execution Plan
func (p *Planner) Plan(capabilityID string, profile model.ExecutionProfile) (*model.ExecutionPlan, error) {
	rc, err := p.resolver.ResolveCapability(capabilityID)
	if err != nil {
		return nil, fmt.Errorf("plan %q: %w", capabilityID, err)
	}

	cap := rc.Capability
	plan := &model.ExecutionPlan{
		CapabilityID:  cap.ID,
		Version:       cap.Version,
		Profile:       string(profile),
		InputSchema:   cap.Inputs,
		OutputSchema:  cap.Outputs,
		Permissions:   cap.Permissions,
		Observability: cap.Observability,
		CreatedAt:     time.Now(),
	}

	// 填充 ResolvedRules
	for _, rule := range rc.ResolvedRules {
		plan.ResolvedRules = append(plan.ResolvedRules, model.ResolvedRule{
			RuleID:      rule.ID,
			Title:       rule.Title,
			Severity:    rule.Severity,
			Description: rule.Description,
		})
	}

	// 单个 Capability 就是单步 Plan
	plan.Steps = []model.PlanStep{
		{
			CapabilityID: cap.ID,
			InputMapping: p.buildInputMapping(cap),
		},
	}

	// 生成 plan_hash
	plan.PlanHash = p.generateHash(plan)

	return plan, nil
}

// PlanWorkflow 展开 Workflow 为多步 Execution Plan
func (p *Planner) PlanWorkflow(workflowID string, profile model.ExecutionProfile) (*model.ExecutionPlan, error) {
	steps, err := p.resolver.ResolveWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("plan workflow %q: %w", workflowID, err)
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("workflow %q has no steps", workflowID)
	}

	plan := &model.ExecutionPlan{
		CapabilityID: workflowID,
		Profile:      string(profile),
		CreatedAt:    time.Now(),
	}

	// 展开每一步
	for _, stepID := range steps {
		rc, err := p.resolver.ResolveCapability(stepID)
		if err != nil {
			return nil, fmt.Errorf("workflow %q step %q: %w", workflowID, stepID, err)
		}

		cap := rc.Capability
		plan.Steps = append(plan.Steps, model.PlanStep{
			CapabilityID: cap.ID,
			InputMapping: p.buildInputMapping(cap),
		})

		for _, rule := range rc.ResolvedRules {
			plan.ResolvedRules = append(plan.ResolvedRules, model.ResolvedRule{
				RuleID:      rule.ID,
				Title:       rule.Title,
				Severity:    rule.Severity,
				Description: rule.Description,
			})
		}

		plan.Permissions = append(plan.Permissions, cap.Permissions...)
	}

	// 去重 permissions
	seen := make(map[string]bool)
	var uniquePerms []string
	for _, p := range plan.Permissions {
		if !seen[p] {
			seen[p] = true
			uniquePerms = append(uniquePerms, p)
		}
	}
	plan.Permissions = uniquePerms

	plan.PlanHash = p.generateHash(plan)

	return plan, nil
}

func (p *Planner) buildInputMapping(cap *model.Capability) map[string]string {
	mapping := make(map[string]string, len(cap.Inputs))
	for _, input := range cap.Inputs {
		mapping[input.Name] = fmt.Sprintf("${%s}", input.Name)
	}
	return mapping
}

func (p *Planner) generateHash(plan *model.ExecutionPlan) string {
	hashInput := fmt.Sprintf("%s:%d:%d:%v", plan.CapabilityID, plan.Version, plan.CreatedAt.UnixNano(), plan.ResolvedRules)
	hash := sha256.Sum256([]byte(hashInput))
	return fmt.Sprintf("%x", hash[:16]) // 前 16 字节作为 plan_hash
}
