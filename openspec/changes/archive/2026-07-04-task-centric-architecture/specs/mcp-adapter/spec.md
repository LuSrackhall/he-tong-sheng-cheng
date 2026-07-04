## ADDED Requirements

### Requirement: MCP Server（Phase 4 — 延迟实现）

MCP Adapter 在 Capability DSL 稳定运行数月后实现。

- MCP MUST 通过 Capability DSL 的 `inputs`/`outputs`/`permissions`/`observability` 自动生成 Tool 定义
- MCP MUST 是 Capability-sandboxed（不能绕过 Knowledge Runtime）
- MCP 的所有执行 MUST 经过完整的 Plan → Guard → Adapter 链路
- MCP 不能直接调用后端 API

#### Scenario: MCP 调用收租金 Tool
- **WHEN** Agent 调用 collect-rent MCP Tool
- **THEN** MCP Server 通过 Knowledge Runtime 生成 Execution Plan
- **THEN** MCP Adapter 执行 Plan，调用后端 Payment API

### Requirement: 安全边界

- MCP Server MUST 使用相同的 Constitution Enforcement Model
- MCP 的 execution_profile 为 `production`
- MCP 的 Pre-execution Guard MUST 为严格模式（非法 Plan 直接拒绝，不降级）

#### Scenario: MCP 拒绝非法请求
- **WHEN** Agent 调用不存在的 Capability
- **THEN** MCP Server 返回工具未找到错误
- **THEN** 不生成任何 Execution Plan
