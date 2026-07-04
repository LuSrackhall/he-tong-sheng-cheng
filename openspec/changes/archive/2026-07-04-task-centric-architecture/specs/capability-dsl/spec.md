## ADDED Requirements

### Requirement: Capability DSL 格式定义

系统使用 YAML 格式定义业务能力（Capability）。每个 Capability 对应一个不可再分的业务操作。

- Capability 文件 MUST 存放在 `knowledge/capabilities/<id>.yaml`
- Capability 文件 MUST 包含 `id`、`version`、`status`、`title`、`goal`、`inputs`、`outputs`、`preconditions`、`effects`、`permissions`、`observability` 字段
- Capability 的 `inputs` 和 `outputs` MUST 引用 `domain` 类型（Contract, Payment, Receipt 等）
- Capability MUST NOT 包含 UI selector、API path、button label 等实现细节
- Capability MUST 带有 `version`（语义版本号）和 `status`（stable/draft/deprecated）

#### Scenario: 完整的 Capability 定义可被解析
- **WHEN** loader 读取一个合法格式的 collect-rent.yaml
- **THEN** 成功解析为 Capability struct，所有字段正确映射

#### Scenario: Capability 包含 UI 细节时被拒绝
- **WHEN** loader 读取一个包含 `steps:` 或 `selector:` 的 YAML
- **THEN** 解析失败，返回 `ErrInvalidCapability`

### Requirement: Workflow DSL 格式定义

Workflow 是多个 Capability 的有序编排，支持条件分支。

- Workflow 文件 MUST 存放在 `knowledge/workflows/<id>.yaml`
- Workflow MUST 包含 `steps` 数组，每个 step 引用一个 Capability `id`
- Step 支持 `if-then-else` 条件分支（condition 引用 Capability 的 observable state）
- Workflow 的 `steps` 构成有向无环图（DAG）

#### Scenario: 简单线性 Workflow 可被展开
- **WHEN** planner 展开一个无分支的 sign-contract workflow
- **THEN** 返回线性 ExecutionPlan steps 序列

#### Scenario: 含条件分支的 Workflow 生成条件 Plan
- **WHEN** planner 展开含 `if: asset.exists` 的 workflow
- **THEN** 生成的 ExecutionPlan 保留条件分支信息

### Requirement: Rule DSL 格式定义

业务规则独立存放，按 `BR-XXX` 编号引用。

- Rule 文件 MUST 存放在 `knowledge/rules/<id>.yaml`
- Rule MUST 包含 `id`、`title`、`severity`（error/warning/info）、`description`、`applies_to`（引用 Capability ID 列表）
- Capability 的 `preconditions` 通过 BR-ID 引用 Rule
- Rule 独立存放，同一 Rule 可被多个 Capability 引用

#### Scenario: Rule 引用可被 Resolver 解析
- **WHEN** resolver 查找 collect-rent 引用的 BR-001
- **THEN** 返回 BR-001 的完整 Rule 对象

#### Scenario: 引用了不存在的 Rule 时报错
- **WHEN** resolver 查找一个 Capability 引用了 BR-999（不存在）
- **THEN** 返回 `ErrRuleNotFound`
