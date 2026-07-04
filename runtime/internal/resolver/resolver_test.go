package resolver

import (
	"testing"

	"asset-leasing-system/runtime/internal/model"
)

func TestResolveCapability(t *testing.T) {
	caps := []*model.Capability{{
		ID:      "test-cap",
		Title:   "Test",
		Preconditions: []string{"BR-001"},
		Dependencies: model.CapabilityDeps{Requires: []string{}},
	}}
	rules := []*model.Rule{{
		ID: "BR-001", Title: "Test Rule", Severity: "error",
	}}

	r := New(caps, rules, nil)
	rc, err := r.ResolveCapability("test-cap")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if len(rc.ResolvedRules) != 1 {
		t.Errorf("expected 1 resolved rule, got %d", len(rc.ResolvedRules))
	}
}

func TestResolveCapabilityNotFound(t *testing.T) {
	r := New(nil, nil, nil)
	_, err := r.ResolveCapability("does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown capability")
	}
}

func TestResolveCapabilityMissingRule(t *testing.T) {
	caps := []*model.Capability{{
		ID: "bad-cap", Preconditions: []string{"BR-999"},
	}}
	r := New(caps, nil, nil)
	_, err := r.ResolveCapability("bad-cap")
	if err == nil {
		t.Fatal("expected error for missing rule")
	}
}

func TestValidateIntegrity(t *testing.T) {
	caps := []*model.Capability{{
		ID: "good", Preconditions: []string{"BR-001"},
		Dependencies: model.CapabilityDeps{Requires: []string{}},
	}, {
		ID: "bad", Preconditions: []string{"BR-999"},
		Dependencies: model.CapabilityDeps{Requires: []string{}},
	}}
	rules := []*model.Rule{{
		ID: "BR-001", Title: "Exists",
	}}

	r := New(caps, rules, nil)
	errs := r.ValidateIntegrity()
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestDetectCycle(t *testing.T) {
	caps := []*model.Capability{
		{ID: "a", Dependencies: model.CapabilityDeps{Requires: []string{"b"}}},
		{ID: "b", Dependencies: model.CapabilityDeps{Requires: []string{"c"}}},
		{ID: "c", Dependencies: model.CapabilityDeps{Requires: []string{"a"}}},
	}

	r := New(caps, nil, nil)
	errs := r.DetectCycle()
	if len(errs) == 0 {
		t.Fatal("expected cycle detection")
	}
}

func TestResolveWorkflow(t *testing.T) {
	caps := []*model.Capability{
		{ID: "step-one", Dependencies: model.CapabilityDeps{Requires: []string{}}},
		{ID: "step-two", Dependencies: model.CapabilityDeps{Requires: []string{}}},
	}
	wfs := []*model.Workflow{{
		ID: "test-wf",
		Steps: []model.WorkflowStep{
			{Capability: "step-one"},
			{Capability: "step-two"},
		},
	}}

	r := New(caps, nil, wfs)
	steps, err := r.ResolveWorkflow("test-wf")
	if err != nil {
		t.Fatalf("resolve workflow failed: %v", err)
	}
	if len(steps) < 2 {
		t.Errorf("expected at least 2 steps, got %d", len(steps))
	}
}
