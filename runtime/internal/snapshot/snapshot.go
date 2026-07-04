package snapshot

import (
	"fmt"
	"time"

	"asset-leasing-system/runtime/internal/model"
	"asset-leasing-system/runtime/internal/planner"
	"asset-leasing-system/runtime/internal/resolver"
)

// Snapshot 负责 Execution Plan 的版本锁定和快照管理
type Snapshot struct {
	resolver *resolver.Resolver
	planner  *planner.Planner
}

func New(res *resolver.Resolver, plan *planner.Planner) *Snapshot {
	return &Snapshot{
		resolver: res,
		planner:  plan,
	}
}

// VersionKey 为版本锁定生成唯一 key
func (s *Snapshot) VersionKey(capabilityID string, version int, profile model.ExecutionProfile) string {
	return fmt.Sprintf("%s:v%d:%s:%d", capabilityID, version, profile, time.Now().UnixMilli())
}

// FreezePlan 从 Execution Plan 创建不可变的版本快照
func (s *Snapshot) FreezePlan(plan *model.ExecutionPlan) *model.ExecutionPlan {
	// plan_hash 已由 planner 分配，此处做二次确认
	if plan.PlanHash == "" {
		plan.PlanHash = "frozen:" + s.VersionKey(plan.CapabilityID, plan.Version, model.ExecutionProfile(plan.Profile))
	}
	return plan
}
