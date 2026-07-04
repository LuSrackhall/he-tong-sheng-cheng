package guard

import (
	"fmt"
	"strings"

	"asset-leasing-system/runtime/internal/model"
)

// ConstitutionGuard 执行 Pre-execution Constitution 验证
type ConstitutionGuard struct {
	knownCapabilities map[string]bool
	strictMode        bool
}

// GuardResult 验证结果
type GuardResult struct {
	Passed     bool
	Violations []*model.ConstitutionViolation
}

func New(knownCaps map[string]bool) *ConstitutionGuard {
	return &ConstitutionGuard{
		knownCapabilities: knownCaps,
		strictMode:        true,
	}
}

// SetStrictMode 设置严格模式
func (g *ConstitutionGuard) SetStrictMode(strict bool) {
	g.strictMode = strict
}

// Check 对 Execution Plan 执行 Constitution 验证
func (g *ConstitutionGuard) Check(plan *model.ExecutionPlan) *GuardResult {
	result := &GuardResult{Passed: true}

	// §1 Knowledge Authority: capability_id 必须在 knowledge/ 中注册
	if g.knownCapabilities != nil {
		if !g.knownCapabilities[plan.CapabilityID] {
			result.Violations = append(result.Violations, &model.ConstitutionViolation{
				Axiom:  "§1",
				Detail: fmt.Sprintf("capability %q is not registered in knowledge layer", plan.CapabilityID),
			})
		}
	}

	// §2 Capability Atomicity: 所有 step 可回溯到 Capability
	for i, step := range plan.Steps {
		if g.knownCapabilities != nil && !g.knownCapabilities[step.CapabilityID] {
			result.Violations = append(result.Violations, &model.ConstitutionViolation{
				Axiom:  "§2",
				Detail: fmt.Sprintf("step %d references unregistered capability %q", i, step.CapabilityID),
			})
		}
	}

	// §3 Plan Immutability: plan_hash 在 emit 前不可为空
	if plan.PlanHash == "" {
		result.Violations = append(result.Violations, &model.ConstitutionViolation{
			Axiom:  "§3",
			Detail: "plan_hash is empty before emit",
		})
	}

	// §4 Separation of Concerns: Plan 不包含业务逻辑执行指令
	if hasBusinessLogic(plan) {
		result.Violations = append(result.Violations, &model.ConstitutionViolation{
			Axiom:  "§4",
			Detail: "plan contains business logic instructions",
		})
	}

	// §9 UI as Derived: Plan 不包含 UI selector 信息
	if hasUISelector(plan) {
		result.Violations = append(result.Violations, &model.ConstitutionViolation{
			Axiom:  "§9",
			Detail: "plan contains UI selector information",
		})
	}

	if len(result.Violations) > 0 {
		if g.strictMode {
			result.Passed = false
		}
		// 非严格模式：记录 violation 但不阻断
	}

	return result
}

// hasBusinessLogic 检查 Plan 是否包含业务逻辑指令
func hasBusinessLogic(plan *model.ExecutionPlan) bool {
	for _, step := range plan.Steps {
		for key := range step.InputMapping {
			if strings.HasPrefix(step.InputMapping[key], "exec:") {
				return true
			}
		}
	}
	return false
}

// hasUISelector 检查 Plan 是否包含 UI selector 信息
func hasUISelector(plan *model.ExecutionPlan) bool {
	for _, step := range plan.Steps {
		if step.InputMapping == nil {
			continue
		}
		for _, val := range step.InputMapping {
			if strings.Contains(val, "#") || strings.Contains(val, ".") || strings.HasPrefix(val, "//") {
				return true
			}
		}
	}
	return false
}
