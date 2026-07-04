## ADDED Requirements

### Requirement: CLI 入口

CLI Adapter 提供三个子命令：

- `kr run <capability> [--inputs...]` — 执行一个 Capability
- `kr plan <capability | workflow>` — 预览 Execution Plan
- `kr explain --trace <trace_id>` — 回溯解释一次 Trace

#### Scenario: kr run 成功执行 Capability
- **WHEN** 执行 `kr run collect-rent --contract 1 --amount 500`
- **THEN** 输出包含 Execution Plan 摘要、执行结果、Trace ID
- **THEN** Trace 文件写入 `.traces/` 目录

#### Scenario: kr plan 预览 Execution Plan
- **WHEN** 执行 `kr plan collect-rent`
- **THEN** 输出包含 plan_hash、steps、前置规则、权限要求
- **THEN** 不执行任何实际操作

### Requirement: kr run 的 Plan 生命周期

- `kr run` MUST 先通过 Runtime 生成 Execution Plan
- Execution Plan 生成后 MUST 通过 Pre-execution Guard 校验 Constitution
- Guard 通过后，CLI 调用后端 API 执行业务操作
- 执行完成后，写入 Execution Trace

#### Scenario: Pre-execution Guard 拒绝非法 Plan
- **WHEN** 执行 `kr run`，但 Runtime 生成违反 Constitution 的 Plan
- **THEN** CLI 输出 Constitution violation 错误，不执行操作，不写入 Trace

### Requirement: kr explain 回溯

- `kr explain` MUST 按 trace_id 读取 Trace 文件
- `kr explain` MUST 输出 Trace 的完整内容（plan_snapshot、steps、observability checks）
- Trace 不存在时 MUST 输出 "Trace not found"

#### Scenario: 回溯成功执行
- **WHEN** 执行 `kr explain --trace abc123` 且该 Trace 存在
- **THEN** 输出 Trace 的完整 JSON

### Requirement: CLI 配置

- CLI MUST 接受 `--knowledge-dir` 参数（默认 `./knowledge`）
- CLI MUST 接受 `--api-base` 参数（后端 API 地址，默认 `http://localhost:8080`）
- CLI MUST 接受 `--validate` 参数（启用/禁用 Pre-execution Guard，默认启用）

#### Scenario: 使用自定义 Knowledge 目录
- **WHEN** 执行 `kr plan collect-rent --knowledge-dir /path/to/knowledge`
- **THEN** Runtime 从指定目录加载 Capability 定义
