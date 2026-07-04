## ADDED Requirements

### Requirement: Trace 数据结构

Trace 是每次 Execution Plan 执行的不可变记录。

- Trace MUST 包含 `identity`（trace_id, capability_id, adapter, plan_version, plan_hash）
- Trace MUST 包含 `context`（user, session, timestamp, environment）
- Trace MUST 包含 `execution_profile`（strict | exploratory | production | debug）
- Trace MUST 包含 `plan_snapshot`（执行时的完整 plan，版本锁定）
- Trace MUST 包含 `steps` 数组（每个 step 包含 intent_step、runtime_step、input、output、duration_ms、error）
- Trace MUST 包含 `observability.checks` 数组（每个 check 包含 rule_id、passed）
- Trace MUST 包含 `determinism`（score 和 factors）
- Trace MUST 包含 `summary`（duration_ms, step_count, error_count）

#### Scenario: 成功执行生成完整 Trace
- **WHEN** CLI Adapter 完成 collect-rent 执行
- **THEN** Trace 包含 identity、context、至少一个 step、observability checks 全部通过

#### Scenario: 失败执行生成含错误信息的 Trace
- **WHEN** CLI Adapter 执行 collect-rent 失败（违反 BR-004）
- **THEN** Trace 的 observability.checks 包含 BR-004: passed=false

### Requirement: Trace 存储

- Trace 文件 MUST 写入 `.traces/YYYY/MM/DD/trace_<hash>.json`
- Trace 文件写入后 MUST 不可变（硬件层面不做强制，但修改后 hash 校验失败）
- Trace 的 plan_hash MUST 与对应的 Execution Plan 一致

#### Scenario: Trace 写入正确路径
- **WHEN** CLI 写入 Trace
- **THEN** 文件位于 `.traces/$(date +%Y/%m/%d)/trace_<hash>.json`

#### Scenario: Trace 文件格式正确
- **WHEN** 读取 Trace 文件
- **THEN** 内容是合法 JSON，可反序列化为 Trace struct

### Requirement: Determinism Score

- Trace MUST 计算 `determinism.score`（0.0–1.0）
- CLI Adapter 的 determinism score MUST 为 0.99
- `factors` 包含 `ui_stability`、`api_consistency`、`agent_variance`

#### Scenario: CLI 执行的 Determinism Score
- **WHEN** CLI 执行完成后生成 Trace
- **THEN** determinism.score = 0.99, agent_variance = 1.0（无 Agent 参与）
