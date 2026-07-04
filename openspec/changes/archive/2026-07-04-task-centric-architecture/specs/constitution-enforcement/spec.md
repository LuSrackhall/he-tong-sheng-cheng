## ADDED Requirements

### Requirement: Pre-execution Guard（Phase 1）

Pre-execution Guard 在 Execution Plan 发出之前验证 Constitution。

- Guard MUST 在 Runtime.Plan() 返回前执行
- Guard MUST 检查 Plan 的以下属性：
  - Plan.capability_id 是否在 knowledge/ 中注册（§1 Knowledge Authority）
  - Plan 的所有 step 是否可回溯到 Capability 或 Workflow（§2 Capability Atomicity）
  - Plan.plan_hash 在 emit 前不可变（§3 Plan Immutability）
  - Plan 是否包含业务逻辑执行指令（§4 Separation of Concerns — 应无）
  - Plan 是否包含 UI selector 信息（§9 UI as Derived — 应无）
- Guard 失败时返回 `ErrConstitutionViolation{axiom, detail}`，不写入 Trace
- Guard 支持宽松模式（`--validate=false`），开发阶段可跳过
- 宽松模式下：violations 仍被记录但执行不被阻断

#### Scenario: 合法 Plan 通过 Guard
- **WHEN** 合法 Plan 进入 Guard
- **THEN** 所有检查通过，Plan 被标记为 emitted

#### Scenario: 非法 Plan 被 Guard 拒绝
- **WHEN** Plan 引用了不存在的 Capability ID
- **THEN** Guard 返回 ErrConstitutionViolation{axiom: "§1"}, Plan 不被 emit

### Requirement: Runtime Guard（Phase 3 — 延迟实现）

- Runtime Guard 的定义和实现在 Phase 3 进行
- Phase 1 的 Guard 实现 MUST 不假设 Runtime Guard 的存在或不存在
- Phase 1 的 Guard 实现 MUST 不与未来 Runtime Guard 冲突

#### Scenario: Phase 1 Guard 不依赖 Runtime Guard
- **WHEN** 审查 Phase 1 Guard 实现代码
- **THEN** 没有引用或依赖 Runtime Guard 相关的类型或接口

### Requirement: Post-execution Audit（Phase 2 — 延迟实现）

- Post-execution Audit 在 CI 集成阶段（Phase 2）引入
- Phase 1 的 Trace 格式 MUST 向后兼容 Post-execution Audit

#### Scenario: Phase 1 Trace 格式兼容
- **WHEN** 审查 Phase 1 Trace 格式
- **THEN** Trace 包含 Post-execution Audit 需要的所有字段（identity, steps, observability）
