package model

import "time"

// ExecutionPlan 是 Knowledge Runtime 和 Adapter 之间的唯一契约
type ExecutionPlan struct {
	CapabilityID  string              `json:"capabilityId"`
	Version       int                 `json:"version"`
	PlanHash      string              `json:"planHash"`
	Profile       string              `json:"executionProfile"` // strict | exploratory | production | debug
	UIMappingVer  string              `json:"uiMappingVersion,omitempty"`

	Steps        []PlanStep           `json:"steps"`
	InputSchema  []CapabilityIO       `json:"inputSchema"`
	OutputSchema []CapabilityIO       `json:"outputSchema"`
	Permissions  []string             `json:"permissions"`
	Observability Observability       `json:"observability"`
	ResolvedRules []ResolvedRule      `json:"resolvedRules"`

	CreatedAt time.Time `json:"createdAt"`
}

type PlanStep struct {
	CapabilityID string                `json:"capabilityId"`
	Condition    *PlanStepCondition    `json:"condition,omitempty"`
	InputMapping map[string]string     `json:"inputMapping"`
}

type PlanStepCondition struct {
	Expression string `json:"expression"`
	ThenStep   int    `json:"thenStep,omitempty"`
	ElseStep   int    `json:"elseStep,omitempty"`
}

type ResolvedRule struct {
	RuleID      string `json:"ruleId"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}
