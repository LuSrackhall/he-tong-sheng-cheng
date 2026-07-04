package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"asset-leasing-system/runtime/internal/model"
)

// Loader 负责从 knowledge/ 目录加载 Capability/Workflow/Rule
type Loader struct {
	knowledgeDir string
}

func New(knowledgeDir string) *Loader {
	return &Loader{knowledgeDir: knowledgeDir}
}

func (l *Loader) LoadAllCapabilities() ([]*model.Capability, error) {
	files, err := l.listYAML("capabilities")
	if err != nil {
		return nil, err
	}
	var caps []*model.Capability
	for _, f := range files {
		cap, err := l.loadCapability(f)
		if err != nil {
			return nil, fmt.Errorf("loading capability %s: %w", f, err)
		}
		caps = append(caps, cap)
	}
	return caps, nil
}

func (l *Loader) LoadCapability(id string) (*model.Capability, error) {
	path := filepath.Join(l.knowledgeDir, "capabilities", id+".yaml")
	return l.loadCapability(path)
}

func (l *Loader) loadCapability(path string) (*model.Capability, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cap model.Capability
	if err := yaml.Unmarshal(data, &cap); err != nil {
		return nil, fmt.Errorf("invalid capability YAML: %w", err)
	}
	if cap.ID == "" {
		return nil, fmt.Errorf("capability missing 'id' field in %s", path)
	}
	if cap.Title == "" {
		return nil, fmt.Errorf("capability %q missing 'title' field", cap.ID)
	}
	return &cap, nil
}

func (l *Loader) LoadAllRules() ([]*model.Rule, error) {
	files, err := l.listYAML("rules")
	if err != nil {
		return nil, err
	}
	var rules []*model.Rule
	for _, f := range files {
		rule, err := l.loadRule(f)
		if err != nil {
			return nil, fmt.Errorf("loading rule %s: %w", f, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (l *Loader) LoadRule(id string) (*model.Rule, error) {
	path := filepath.Join(l.knowledgeDir, "rules", id+".yaml")
	return l.loadRule(path)
}

func (l *Loader) loadRule(path string) (*model.Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rule model.Rule
	if err := yaml.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("invalid rule YAML: %w", err)
	}
	if rule.ID == "" {
		return nil, fmt.Errorf("rule missing 'id' field in %s", path)
	}
	return &rule, nil
}

func (l *Loader) LoadAllWorkflows() ([]*model.Workflow, error) {
	files, err := l.listYAML("workflows")
	if err != nil {
		return nil, err
	}
	var wfs []*model.Workflow
	for _, f := range files {
		wf, err := l.loadWorkflow(f)
		if err != nil {
			return nil, fmt.Errorf("loading workflow %s: %w", f, err)
		}
		wfs = append(wfs, wf)
	}
	return wfs, nil
}

func (l *Loader) LoadWorkflow(id string) (*model.Workflow, error) {
	path := filepath.Join(l.knowledgeDir, "workflows", id+".yaml")
	return l.loadWorkflow(path)
}

func (l *Loader) loadWorkflow(path string) (*model.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wf model.Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("invalid workflow YAML: %w", err)
	}
	if wf.ID == "" {
		return nil, fmt.Errorf("workflow missing 'id' field in %s", path)
	}
	return &wf, nil
}

func (l *Loader) listYAML(subdir string) ([]string, error) {
	dir := filepath.Join(l.knowledgeDir, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".yaml" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}
