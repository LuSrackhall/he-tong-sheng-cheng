package model

// Workflow 定义多个 Capability 的有序编排
type Workflow struct {
	ID    string `yaml:"id" json:"id"`
	Title string `yaml:"title" json:"title"`

	Metadata CapabilityMetadata `yaml:"metadata" json:"metadata"`

	Steps []WorkflowStep `yaml:"steps" json:"steps"`
}

type WorkflowStep struct {
	Capability string          `yaml:"task,omitempty" json:"task,omitempty"`
	Condition  *WorkflowCondition `yaml:"if,omitempty" json:"if,omitempty"`
	Then       string          `yaml:"then,omitempty" json:"then,omitempty"`
	Else       string          `yaml:"else,omitempty" json:"else,omitempty"`
}

type WorkflowCondition struct {
	Condition string `yaml:"condition" json:"condition"`
}
