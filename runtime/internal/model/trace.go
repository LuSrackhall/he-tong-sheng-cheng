package model

import "time"

// Trace 是每次 Execution Plan 执行的不可变记录
type Trace struct {
	Identity  TraceIdentity   `json:"identity"`
	Context   TraceContext    `json:"context"`
	Profile   string          `json:"executionProfile"`

	PlanSnapshot ExecutionPlan `json:"planSnapshot"`

	Steps        []TraceStep   `json:"steps"`
	Observability TraceObs     `json:"observability"`
	Determinism  Determinism   `json:"determinism"`
	Summary      TraceSummary  `json:"summary"`
	Feedback     *Feedback     `json:"feedback,omitempty"`
}

type TraceIdentity struct {
	TraceID        string `json:"traceId"`
	CapabilityID   string `json:"capabilityId"`
	Adapter        string `json:"adapter"`
	PlanVersion    int    `json:"planVersion"`
	PlanHash       string `json:"planHash"`
}

type TraceContext struct {
	User        string    `json:"user"`
	Session     string    `json:"session"`
	Timestamp   time.Time `json:"timestamp"`
	Environment string    `json:"environment"`
}

type TraceStep struct {
	IntentStep  string         `json:"intentStep"`
	RuntimeStep *RuntimeStepDetail `json:"runtimeStep,omitempty"`
	Input       map[string]any `json:"input"`
	Output      map[string]any `json:"output"`
	DurationMs  int64          `json:"durationMs"`
	Error       *StepError     `json:"error,omitempty"`
}

type RuntimeStepDetail struct {
	Selector string `json:"selector,omitempty"`
	Action   string `json:"action"`
}

type StepError struct {
	RuleID  string `json:"ruleId,omitempty"`
	Message string `json:"message"`
}

type TraceObs struct {
	Success []string       `json:"success"`
	Failure []string       `json:"failure"`
	Checks  []ObservabilityCheck `json:"checks"`
}

type ObservabilityCheck struct {
	RuleID string `json:"ruleId"`
	Passed bool   `json:"passed"`
}

type Determinism struct {
	Score   float64           `json:"score"`
	Factors map[string]float64 `json:"factors"`
}

type TraceSummary struct {
	DurationMs int64 `json:"durationMs"`
	StepCount  int   `json:"stepCount"`
	ErrorCount int   `json:"errorCount"`
}

type Feedback struct {
	Adapter       string `json:"adapter"`
	Type          string `json:"type"`
	Severity      string `json:"severity"`
	Capability    string `json:"capability"`
	Scenario      string `json:"scenario"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Recommendation string `json:"recommendation"`
}
