package model

// Rule 定义业务规则
type Rule struct {
	ID          string   `yaml:"id" json:"id"`
	Title       string   `yaml:"title" json:"title"`
	Severity    string   `yaml:"severity" json:"severity"` // error | warning | info
	Description string   `yaml:"description" json:"description"`
	AppliesTo   []string `yaml:"applies_to" json:"applies_to"`
}
