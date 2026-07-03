## Why

目前系统业务模型成熟、架构完整，但缺乏 E2E 测试、用户体验验证以及面向 Agent 的运行时接口。核心痛点不是功能缺失，而是缺乏对"业务任务"的形式化定义和执行质量的系统性保障。本 change 引入 Task-Centric Architecture（TCA），将业务能力抽象为领域层一等公民，建立 Knowledge Runtime + Runtime Adapters + CI 一致性验证 + Constitution 治理的完整体系。

## What Changes

- **新增 Capability DSL** — 以 `knowledge/capabilities/` 为中心的 YAML 格式业务能力定义，取代无结构的任务描述
- **新增 Knowledge Runtime** — 五层运行时（Storage → Parser → Resolver → Planner → Query），生成不可变的 Execution Plan
- **新增 Execution Plan 契约** — Runtime 和 Adapter 之间的唯一通信协议，plan_hash 锁定，不可修改
- **新增 CLI Adapter** — `kr run/plan/explain` 命令行工具，用于调试和验证 Knowledge Runtime
- **新增 Runtime Adapters 体系** — Browser（Playwright 语义回放）/ Exploration（Agent UX 审计）/ MCP（生产运行时）
- **新增 Execution Trace 系统** — 不可变执行轨迹，作为系统唯一 truth source for behavioral learning
- **新增 CI 语义验证系统** — execution graph 变化影响分析，非传统测试运行器
- **新增 Constitution 治理层** — 13 条不可违反公理 + 三层 Enforcement Gate
- **新增 `knowledge/` 顶层目录** — 业务知识层（非 `internal/`，非 `openspec/`），包含 capabilities / workflows / rules / glossary

## Capabilities

### New Capabilities

- `capability-dsl`: 业务能力定义格式（Capability / Workflow / Rule YAML schema），语言无关，不依赖 UI/API 细节
- `knowledge-runtime`: 五层运行时（Storage → Parser → Resolver → Planner → Query），加载 capability、解析引用、生成 Execution Plan
- `cli-adapter`: Knowledge Runtime 的 CLI 接口，支持 `kr run`（执行）/ `kr plan`（预览）/ `kr explain`（回溯）
- `execution-trace`: 不可变执行轨迹系统，包含 plan_snapshot、step 序列、observability checks、determinism score
- `browser-adapter`: Reference Browser Adapter，Capability → UI Action Plan 映射 + Playwright 语义回放
- `exploration-adapter`: Agent 驱动的 UX 审计 / 混沌测试 / 可发现性测试
- `ci-replay`: CI 语义验证系统，exact replay + semantic replay + 三维 drift 分析
- `constitution-enforcement`: Constitution 三层 Gate（Pre-execution / Runtime / Post-execution）
- `mcp-adapter`: MCP Server 适配器（Phase 4）

### Modified Capabilities

（无现有 spec 被修改。TCA 是新增层，不改变现有业务语义。）

## Impact

- **新增 `knowledge/` 顶层目录** — 与 `internal/`、`frontend/` 同级，不属于 Go 模块
- **新增 `runtime/` 目录** — Knowledge Runtime 的 Go 实现
- **新增 `cli/` 目录** — CLI Adapter 入口
- **新增 `adapters/` 目录** — Browser / Exploration / MCP 适配器
- **新增 `ci/` 目录** — CI 集成工具
- **新增 `system/constitution.md`** — 全局不可变宪法
- **新增 `.traces/` 目录** — Trace 存档
- **Phase 1 不修改** 现有 Go 后端分层（domain/repository/transport）、现有 UI、现有 OpenSpec 工作流
- **新增依赖** 仅 Playwright（Phase 2），不引入新后端依赖
