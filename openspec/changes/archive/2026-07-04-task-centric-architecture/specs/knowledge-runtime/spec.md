## ADDED Requirements

### Requirement: Knowledge Runtime 架构

Knowledge Runtime 是 TCA 的核心引擎，负责加载、解析、规划、查询 Capability。Runtime 不执行业务逻辑。

- Runtime MUST 提供五层处理管线：Storage → Parser → Resolver → Planner → Query
- Runtime MUST 接受 `knowledge/` 目录路径作为输入
- Runtime MUST 在启动时验证所有 Capability/Workflow/Rule 的引用完整性
- Runtime MUST 检测 Workflow DAG 环路，有环时报错

#### Scenario: Runtime 启动时校验引用完整性
- **WHEN** runtime 启动，加载 knowledge/ 目录
- **THEN** 检查所有 BR-ID 引用存在、所有 workflow steps 引用的 capability 存在
- **THEN** DAG 无环验证通过

#### Scenario: 引用不完整时报错
- **WHEN** runtime 启动，发现某个 Capability 引用了不存在的 BR-ID
- **THEN** 返回引用错误列表，启动失败

### Requirement: Loader（Storage 层）

- Loader MUST 递归扫描 `knowledge/capabilities/`、`knowledge/workflows/`、`knowledge/rules/`
- Loader MUST 支持 YAML 解析
- Loader MUST 支持按 `id` 查询单个 Capability，支持按 `tag` 批量查询
- Loader MUST 支持 Capability 按 `status` 过滤（stable/draft/deprecated）

#### Scenario: 按 ID 加载 Capability
- **WHEN** loader.LoadCapability("collect-rent") 被调用
- **THEN** 返回 collect-rent 的完整 Capability struct

#### Scenario: 按 Tag 查询
- **WHEN** loader.FindCapabilities({tags: ["rent"]}) 被调用
- **THEN** 返回所有包含 rent tag 的 Capability 列表

### Requirement: Resolver（引用解析层）

- Resolver MUST 解析 Capability 的 `preconditions` 中的 BR-ID 引用
- Resolver MUST 解析 Workflow 中每个 step 引用的 Capability
- Resolver MUST 构建 Capability 的完整依赖 DAG（通过 `dependencies.requires` 字段）

#### Scenario: 解析 Capability 的 BR 引用
- **WHEN** resolver.Resolve(capability) 被调用
- **THEN** 返回 ResolvedCapability，其中 preconditions 包含 Rule 对象而非 BR-ID

#### Scenario: 构建依赖 DAG
- **WHEN** resolver.GetDependencyGraph("collect-rent") 被调用
- **THEN** 返回包含 collect-rent 及其所有依赖（ensure-contract-active 等）的 DAG

### Requirement: Planner（规划层）

- Planner MUST 将单个 Capability 展开为 Execution Plan
- Planner MUST 将 Workflow 展开为有序的 Capability 步骤序列
- Planner MUST 处理 Workflow 中的条件分支（`if-then-else`）
- Planner MUST 在生成的 Execution Plan 上分配 `plan_hash`

#### Scenario: Capability 生成 Execution Plan
- **WHEN** planner.Plan("collect-rent") 被调用
- **THEN** 返回 ExecutionPlan：steps = [collect-rent], plan_hash 非空, resolved rules 齐全

#### Scenario: Workflow 展开为多步骤 Plan
- **WHEN** planner.PlanWorkflow("sign-new-contract") 被调用
- **THEN** 返回 ExecutionPlan：steps 为展开后的多步骤序列

### Requirement: Cache（缓存层）

- Cache MUST 支持 Capability、Rule、Workflow、Execution Plan 四种类型的缓存
- Cache MUST 提供失效机制（按 ID 或按 Tag 失效）
- Cache 默认实现为内存缓存（sync.Map），接口可替换

#### Scenario: Capability 缓存命中
- **WHEN** cache.Get("capability:collect-rent") 被调用且存在缓存
- **THEN** 返回缓存中的 Capability struct

#### Scenario: 缓存失效后重新加载
- **WHEN** cache.Invalidate("capability:collect-rent") 后再次 Get
- **THEN** Loader 重新从文件加载
