## ADDED Requirements

### Requirement: Runtime Validation（Phase 2）

CI 在代码变更后执行 Knowledge Runtime 的完整性检查。

- CI MUST 验证所有 Capability/Workflow/Rule 的引用完整性
- CI MUST 检测 DAG 环路
- CI MUST 验证 Constitution §2（所有 execution 必须从 Capability 出发）
- Validation 失败时 CI 阻断（blocking gate）

#### Scenario: 引用不完整时 CI 阻断
- **WHEN** CI 发现某个 Capability 引用了不存在的 BR-ID
- **THEN** CI 输出 validation error，pipeline 标记为失败

### Requirement: 语义回放（Semantic Replay，Phase 2）

- CI 对不同 Adapter 使用不同的回放策略：
  - CLI Trace → Exact Replay（input/output MUST 一致）
  - Browser Trace → Semantic Replay（observability MUST 一致，steps 可以不同）
- Exact Replay 失败时 CI 阻断
- Semantic Replay 失败时 CI 不阻断但记录 drift

#### Scenario: CLI Trace Exact Replay 成功
- **WHEN** CI replay 一个 CLI 产生的 Trace
- **THEN** 验证 input/output 完全匹配，一致性 > 0.95

#### Scenario: Browser Trace Semantic Replay 发现 UI 变化
- **WHEN** CI replay 一个 Browser 产生的 Trace
- **THEN** 验证 observability checks 全部通过
- **THEN** UI structural drift 被标记但不阻断

### Requirement: 三维 Drift 评分

- CI MUST 计算三个维度的 drift score：
  - `structural`（steps 一致性）
  - `semantic`（observability 一致性）
  - `ui`（UI Action Plan 变化）
- Drift 基于时间窗口评估（rolling window of traces），非单次执行
- CI 在**持续 drift 恶化趋势**时阻断（sustained degradation），而非单次波动

#### Scenario: 单次 UI 变化不阻断
- **WHEN** 一次 UI 重构导致单次 drift spike
- **THEN** CI 记录、告警但不阻断

#### Scenario: UI drift 持续下降触发阻断
- **WHEN** 连续 5 次 CI 运行中 ui drift 从 0.9 持续降至 0.5
- **THEN** CI 阻断并触发 investigation
