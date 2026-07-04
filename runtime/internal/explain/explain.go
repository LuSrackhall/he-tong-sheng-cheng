package explain

import (
	"encoding/json"
	"fmt"
	"strings"

	"asset-leasing-system/runtime/internal/trace"
)

// Explain 负责回溯解释 Trace
type Explain struct {
	reader *trace.Reader
}

func New(reader *trace.Reader) *Explain {
	return &Explain{reader: reader}
}

// ByTraceID 按 trace_id 解释一次执行
func (e *Explain) ByTraceID(traceID string) (string, error) {
	tr, path, err := e.reader.ReadByID(traceID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Trace: %s\n", tr.Identity.TraceID))
	b.WriteString(fmt.Sprintf("File: %s\n", path))
	b.WriteString(fmt.Sprintf("Capability: %s (v%d)\n", tr.Identity.CapabilityID, tr.Identity.PlanVersion))
	b.WriteString(fmt.Sprintf("Adapter: %s\n", tr.Identity.Adapter))
	b.WriteString(fmt.Sprintf("Profile: %s\n", tr.Profile))
	b.WriteString(fmt.Sprintf("User: %s\n", tr.Context.User))
	b.WriteString(fmt.Sprintf("Timestamp: %s\n", tr.Context.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Duration: %dms\n", tr.Summary.DurationMs))
	b.WriteString(fmt.Sprintf("Steps: %d (errors: %d)\n", tr.Summary.StepCount, tr.Summary.ErrorCount))
	b.WriteString(fmt.Sprintf("Determinism: %.2f\n", tr.Determinism.Score))
	b.WriteString("\n")

	b.WriteString("Steps:\n")
	for i, step := range tr.Steps {
		b.WriteString(fmt.Sprintf("  %d. %s (%dms)\n", i+1, step.IntentStep, step.DurationMs))
		if step.Error != nil {
			b.WriteString(fmt.Sprintf("     ERROR: %s\n", step.Error.Message))
		}
	}

	b.WriteString("\nObservability Checks:\n")
	for _, check := range tr.Observability.Checks {
		status := "✓"
		if !check.Passed {
			status = "✗"
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", status, check.RuleID))
	}

	return b.String(), nil
}

// JSON 以 JSON 格式输出 Trace
func (e *Explain) JSON(traceID string) (string, error) {
	tr, _, err := e.reader.ReadByID(traceID)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(tr, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
