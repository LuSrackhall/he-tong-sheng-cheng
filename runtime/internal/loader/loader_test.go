package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCapability(t *testing.T) {
	// Create temp knowledge dir
	tmpDir := t.TempDir()
	capsDir := filepath.Join(tmpDir, "capabilities")
	os.MkdirAll(capsDir, 0755)

	yamlContent := `
id: test-cap
version: 1
status: stable
title: Test Capability
goal: Test that loading works
inputs:
  - { name: input1, domain: String }
outputs:
  - { name: output1, domain: String }
preconditions:
  - BR-001
effects:
  sync:
    creates: ["Result"]
permissions:
  - test:run
dependencies:
  requires: []
observability:
  success: ["Done"]
  failure: ["Fail"]
`
	os.WriteFile(filepath.Join(capsDir, "test-cap.yaml"), []byte(yamlContent), 0644)

	l := New(tmpDir)
	cap, err := l.LoadCapability("test-cap")
	if err != nil {
		t.Fatalf("failed to load capability: %v", err)
	}
	if cap.ID != "test-cap" {
		t.Errorf("expected test-cap, got %s", cap.ID)
	}
	if cap.Title != "Test Capability" {
		t.Errorf("expected 'Test Capability', got %s", cap.Title)
	}
	if len(cap.Inputs) != 1 {
		t.Errorf("expected 1 input, got %d", len(cap.Inputs))
	}
}

func TestLoadCapabilityMissingID(t *testing.T) {
	tmpDir := t.TempDir()
	capsDir := filepath.Join(tmpDir, "capabilities")
	os.MkdirAll(capsDir, 0755)

	os.WriteFile(filepath.Join(capsDir, "bad.yaml"), []byte("title: No ID"), 0644)

	l := New(tmpDir)
	_, err := l.LoadCapability("bad")
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestLoadRule(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")
	os.MkdirAll(rulesDir, 0755)

	os.WriteFile(filepath.Join(rulesDir, "BR-999.yaml"), []byte(`
id: BR-999
title: Test Rule
severity: error
description: A test rule
applies_to: [test-cap]
`), 0644)

	l := New(tmpDir)
	rule, err := l.LoadRule("BR-999")
	if err != nil {
		t.Fatalf("failed to load rule: %v", err)
	}
	if rule.ID != "BR-999" {
		t.Errorf("expected BR-999, got %s", rule.ID)
	}
}

func TestLoadAllCapabilities(t *testing.T) {
	tmpDir := t.TempDir()
	capsDir := filepath.Join(tmpDir, "capabilities")
	os.MkdirAll(capsDir, 0755)

	os.WriteFile(filepath.Join(capsDir, "a.yaml"), []byte("id: cap-a\nversion: 1\ntitle: A\ngoal: Test\ninputs: []\noutputs: []\npreconditions: []\neffects:\n  sync:\n    creates: []\npermissions: []\ndependencies:\n  requires: []\nobservability:\n  success: []\n  failure: []\n"), 0644)
	os.WriteFile(filepath.Join(capsDir, "b.yaml"), []byte("id: cap-b\nversion: 1\ntitle: B\ngoal: Test\ninputs: []\noutputs: []\npreconditions: []\neffects:\n  sync:\n    creates: []\npermissions: []\ndependencies:\n  requires: []\nobservability:\n  success: []\n  failure: []\n"), 0644)

	l := New(tmpDir)
	caps, err := l.LoadAllCapabilities()
	if err != nil {
		t.Fatalf("failed to load capabilities: %v", err)
	}
	if len(caps) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(caps))
	}
}

func TestLoadWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	wfDir := filepath.Join(tmpDir, "workflows")
	os.MkdirAll(wfDir, 0755)

	os.WriteFile(filepath.Join(wfDir, "test-wf.yaml"), []byte(`
id: test-wf
title: Test Workflow
steps:
  - task: step-one
  - task: step-two
`), 0644)

	l := New(tmpDir)
	wf, err := l.LoadWorkflow("test-wf")
	if err != nil {
		t.Fatalf("failed to load workflow: %v", err)
	}
	if wf.ID != "test-wf" {
		t.Errorf("expected test-wf, got %s", wf.ID)
	}
	if len(wf.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(wf.Steps))
	}
}
