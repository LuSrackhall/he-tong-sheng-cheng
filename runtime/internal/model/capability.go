package model

// Capability 定义系统提供的业务能力
type Capability struct {
	ID      string `yaml:"id" json:"id"`
	Version int    `yaml:"version" json:"version"`
	Status  string `yaml:"status" json:"status"` // stable | draft | deprecated
	Since   string `yaml:"since" json:"since"`

	Metadata CapabilityMetadata `yaml:"metadata" json:"metadata"`

	Title       string `yaml:"title" json:"title"`
	Goal        string `yaml:"goal" json:"goal"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	Inputs  []CapabilityIO `yaml:"inputs" json:"inputs"`
	Outputs []CapabilityIO `yaml:"outputs" json:"outputs"`

	Preconditions []string           `yaml:"preconditions" json:"preconditions"`
	Effects       CapabilityEffects  `yaml:"effects" json:"effects"`
	Permissions   []string           `yaml:"permissions" json:"permissions"`
	Dependencies  CapabilityDeps     `yaml:"dependencies" json:"dependencies"`
	Observability Observability      `yaml:"observability" json:"observability"`
}

type CapabilityMetadata struct {
	Tags     []string `yaml:"tags" json:"tags"`
	Owner    string   `yaml:"owner" json:"owner"`
	Priority string   `yaml:"priority" json:"priority"`
}

type CapabilityIO struct {
	Name     string `yaml:"name" json:"name"`
	Domain   string `yaml:"domain" json:"domain"`
	Min      any    `yaml:"min,omitempty" json:"min,omitempty"`
	Max      any    `yaml:"max,omitempty" json:"max,omitempty"`
	Enum     []any  `yaml:"enum,omitempty" json:"enum,omitempty"`
	Required *bool  `yaml:"required,omitempty" json:"required,omitempty"`
}

type CapabilityEffects struct {
	Sync  SyncEffects  `yaml:"sync" json:"sync"`
	Async AsyncEffects `yaml:"async" json:"async"`
}

type SyncEffects struct {
	Creates []string `yaml:"creates" json:"creates"`
	Updates []string `yaml:"updates" json:"updates"`
}

type AsyncEffects struct {
	Emits []string `yaml:"emits" json:"emits"`
}

type CapabilityDeps struct {
	Requires []string `yaml:"requires" json:"requires"`
	Triggers []string `yaml:"triggers" json:"triggers"`
}

type Observability struct {
	Success []string `yaml:"success" json:"success"`
	Failure []string `yaml:"failure" json:"failure"`
}
