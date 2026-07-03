package planner

import (
	"testing"

	"asset-leasing-system/runtime/internal/model"
	"asset-leasing-system/runtime/internal/resolver"
)

func TestPlanCapability(t *testing.T) {
	caps := []*model.Capability{{
		ID:      "test-cap",
		Version: 1,
		Title:   "Test",
		Inputs:  []model.CapabilityIO{{Name: "input1", Domain: "String"}},
		Outputs: []model.CapabilityIO{{Name: "output1", Domain: "String"}},
		Preconditions: []string{"BR-001"},
		Permissions:   []string{"test:run"},
		Dependencies:  model.CapabilityDeps{Requires: []string{}},
	}}
	rules := []*model.Rule{{
		ID: "BR-001", Title: "Test", Severity: "error",
	}}

	res := resolver.New(caps, rules, nil)
	p := New(res)

	plan, err := p.Plan("test-cap", model.ProfileStrict)
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if plan.PlanHash == "" {
		t.Fatal("expected non-empty plan_hash")
	}
	if len(plan.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(plan.Steps))
	}
	if len(plan.ResolvedRules) != 1 {
		t.Errorf("expected 1 resolved rule, got %d", len(plan.ResolvedRules))
	}
}

func TestPlanCapabilityNotFound(t *testing.T) {
	res := resolver.New(nil, nil, nil)
	p := New(res)

	_, err := p.Plan("does-not-exist", model.ProfileStrict)
	if err == nil {
		t.Fatal("expected error for unknown capability")
	}
}

func TestPlanWorkflow(t *testing.T) {
	caps := []*model.Capability{
		{ID: "step1", Dependencies: model.CapabilityDeps{Requires: []string{}}},
		{ID: "step2", Dependencies: model.CapabilityDeps{Requires: []string{}}},
	}
	wfs := []*model.Workflow{{
		ID: "test-wf",
		Steps: []model.WorkflowStep{
			{Capability: "step1"},
			{Capability: "step2"},
		},
	}}

	res := resolver.New(caps, nil, wfs)
	p := New(res)

	plan, err := p.PlanWorkflow("test-wf", model.ProfileStrict)
	if err != nil {
		t.Fatalf("plan workflow failed: %v", err)
	}
	if len(plan.Steps) < 2 {
		t.Errorf("expected at least 2 steps, got %d", len(plan.Steps))
	}
	if plan.PlanHash == "" {
		t.Fatal("expected non-empty plan_hash")
	}
}

func TestPlanProfileIsSet(t *testing.T) {
	caps := []*model.Capability{{
		ID: "test", Version: 1, Title: "Test",
		Inputs:  []model.CapabilityIO{},
		Outputs: []model.CapabilityIO{},
		Preconditions: []string{},
		Permissions:   []string{},
		Dependencies:  model.CapabilityDeps{Requires: []string{}},
	}}

	res := resolver.New(caps, nil, nil)
	p := New(res)

	plan, _ := p.Plan("test", model.ProfileDebug)
	if plan.Profile != string(model.ProfileDebug) {
		t.Errorf("expected debug profile, got %s", plan.Profile)
	}
}
