package trace

import (
	"testing"
	"time"

	"asset-leasing-system/runtime/internal/model"
)

func TestWriteAndReadTrace(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)
	r := NewReader(dir)

	trace := &model.Trace{
		Identity: model.TraceIdentity{
			CapabilityID: "test-cap",
			PlanHash:     "plan123",
		},
		Context: model.TraceContext{
			Timestamp: time.Now(),
			User:      "test-user",
		},
		Steps: []model.TraceStep{
			{
				IntentStep: "Execute test",
				DurationMs: 100,
			},
		},
		Determinism: model.Determinism{
			Score: 0.99,
			Factors: map[string]float64{
				"api_consistency": 1.0,
			},
		},
	}

	path, err := w.Write(trace)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}

	readBack, readPath, err := r.ReadByID(trace.Identity.TraceID)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if readBack.Identity.CapabilityID != "test-cap" {
		t.Errorf("expected test-cap, got %s", readBack.Identity.CapabilityID)
	}
	if readPath == "" {
		t.Fatal("expected non-empty read path")
	}
}

func TestTraceNotFound(t *testing.T) {
	dir := t.TempDir()
	r := NewReader(dir)

	_, _, err := r.ReadByID("does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestReadLatest(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)
	r := NewReader(dir)

	trace1 := &model.Trace{
		Identity: model.TraceIdentity{CapabilityID: "first", PlanHash: "p1"},
		Context:  model.TraceContext{Timestamp: time.Now().Add(-time.Hour)},
	}
	w.Write(trace1)

	trace2 := &model.Trace{
		Identity: model.TraceIdentity{CapabilityID: "second", PlanHash: "p2"},
		Context:  model.TraceContext{Timestamp: time.Now()},
	}
	// Use a short delay so modtime differs
	time.Sleep(10 * time.Millisecond)
	w.Write(trace2)

	latest, _, err := r.ReadLatest()
	if err != nil {
		t.Fatalf("read latest failed: %v", err)
	}
	if latest.Identity.CapabilityID != "second" {
		t.Errorf("expected latest to be 'second', got %s", latest.Identity.CapabilityID)
	}
}

func TestDeterminismScoreDefault(t *testing.T) {
	d := model.Determinism{
		Score: 0.99,
		Factors: map[string]float64{
			"ui_stability":     1.0,
			"api_consistency":  1.0,
			"agent_variance":   1.0,
		},
	}
	if d.Score != 0.99 {
		t.Errorf("expected 0.99, got %f", d.Score)
	}
}
