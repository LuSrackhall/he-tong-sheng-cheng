package guard

import (
	"testing"

	"asset-leasing-system/runtime/internal/model"
)

func TestGuardValidPlan(t *testing.T) {
	known := map[string]bool{"test-cap": true}
	g := New(known)

	plan := &model.ExecutionPlan{
		CapabilityID: "test-cap",
		PlanHash:     "abc123",
		Steps: []model.PlanStep{
			{CapabilityID: "test-cap", InputMapping: map[string]string{"input": "${input}"}},
		},
	}

	result := g.Check(plan)
	if !result.Passed {
		t.Fatalf("expected valid plan to pass, got violations: %v", result.Violations)
	}
}

func TestGuardMissingCapability(t *testing.T) {
	known := map[string]bool{"real-cap": true}
	g := New(known)

	plan := &model.ExecutionPlan{
		CapabilityID: "fake-cap",
		PlanHash:     "abc123",
	}

	result := g.Check(plan)
	if result.Passed {
		t.Fatal("expected violation for unregistered capability")
	}
	if result.Violations[0].Axiom != "§1" {
		t.Errorf("expected §1 violation, got %s", result.Violations[0].Axiom)
	}
}

func TestGuardEmptyPlanHash(t *testing.T) {
	known := map[string]bool{"test-cap": true}
	g := New(known)

	plan := &model.ExecutionPlan{
		CapabilityID: "test-cap",
		PlanHash:     "",
	}

	result := g.Check(plan)
	if result.Passed {
		t.Fatal("expected violation for empty plan_hash")
	}
	found := false
	for _, v := range result.Violations {
		if v.Axiom == "§3" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected §3 violation")
	}
}

func TestGuardBusinesLogic(t *testing.T) {
	known := map[string]bool{"test-cap": true}
	g := New(known)

	plan := &model.ExecutionPlan{
		CapabilityID: "test-cap",
		PlanHash:     "abc123",
		Steps: []model.PlanStep{
			{CapabilityID: "test-cap", InputMapping: map[string]string{"cmd": "exec:DELETE FROM payments"}},
		},
	}

	result := g.Check(plan)
	if result.Passed {
		t.Fatal("expected violation for business logic")
	}
	if result.Violations[0].Axiom != "§4" {
		t.Errorf("expected §4 violation, got %s", result.Violations[0].Axiom)
	}
}

func TestGuardUISelector(t *testing.T) {
	known := map[string]bool{"test-cap": true}
	g := New(known)

	plan := &model.ExecutionPlan{
		CapabilityID: "test-cap",
		PlanHash:     "abc123",
		Steps: []model.PlanStep{
			{CapabilityID: "test-cap", InputMapping: map[string]string{"sel": "#amount-input"}},
		},
	}

	result := g.Check(plan)
	if result.Passed {
		t.Fatal("expected violation for UI selector")
	}
	hasAxiom9 := false
	for _, v := range result.Violations {
		if v.Axiom == "§9" {
			hasAxiom9 = true
			break
		}
	}
	if !hasAxiom9 {
		t.Error("expected §9 violation")
	}
}

func TestGuardLooseMode(t *testing.T) {
	known := map[string]bool{"real-cap": true}
	g := New(known)
	g.SetStrictMode(false)

	plan := &model.ExecutionPlan{
		CapabilityID: "fake-cap",
		PlanHash:     "",
	}

	result := g.Check(plan)
	if result.Passed {
		t.Fatal("loose mode bypasses guard? it should still check but we test different behavior")
	}
}

func TestGuardMultipleViolations(t *testing.T) {
	known := map[string]bool{}
	g := New(known)

	plan := &model.ExecutionPlan{
		CapabilityID: "fake-cap",
		PlanHash:     "",
		Steps: []model.PlanStep{
			{CapabilityID: "unknown-step"},
		},
	}

	result := g.Check(plan)
	if result.Passed {
		t.Fatal("expected multiple violations")
	}
	if len(result.Violations) < 2 {
		t.Errorf("expected at least 2 violations, got %d", len(result.Violations))
	}
}
