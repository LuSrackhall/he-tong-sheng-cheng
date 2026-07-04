package model

// ResolvedCapability 是引用解析后的 Capability（Rule ID 已替换为 Rule 对象）
type ResolvedCapability struct {
	Capability    *Capability
	ResolvedRules []*Rule
	Dependencies  []string // 解析后的依赖列表
	DAG           map[string][]string
}

// ExecutionProfile 执行模式
type ExecutionProfile string

const (
	ProfileStrict      ExecutionProfile = "strict"
	ProfileExploratory ExecutionProfile = "exploratory"
	ProfileProduction  ExecutionProfile = "production"
	ProfileDebug       ExecutionProfile = "debug"
)

// ConstitutionViolation 宪法违反错误
type ConstitutionViolation struct {
	Axiom  string `json:"axiom"`
	Detail string `json:"detail"`
}

func (v *ConstitutionViolation) Error() string {
	return "Constitution violation §" + v.Axiom + ": " + v.Detail
}
