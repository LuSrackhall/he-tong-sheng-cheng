## Context

TCA（Task-Centric Architecture）已完成完整的系统架构设计（brainstorm-spec.md，580 行）。涵盖 9 层架构：Knowledge Runtime → Execution Plan → Runtime Adapters → Trace → CI → Feedback → Constitution。

本 design.md 聚焦 **Phase 1 Kernel** 的工程实现细节。Phase 1 范围：

- `knowledge/` — Capability DSL 首批定义
- `runtime/` — Knowledge Runtime（Storage → Parser → Resolver → Planner → Query）
- `cli/` — CLI Adapter（kr run / kr plan / kr explain）
- `system/constitution.md` — 宪法全局定义（已完成）

Phase 2+（Browser Adapter、CI、Exploration、MCP）不在本 design 范围内。

## Goals / Non-Goals

**Goals:**

1. Knowledge Runtime 可加载、解析、缓存 Capability/Workflow/Rule YAML
2. Resolver 可解析引用关系（BR-ID → Rule、workflow → capability DAG）
3. Planner 可展开 workflow 为 Execution Plan（含条件分支处理）
4. Execution Plan 生成时分配 `plan_hash`，不可修改
5. CLI Adapter 支持 `kr run`（执行）、`kr plan`（预览）、`kr explain`（回溯 Trace）
6. Pre-execution Guard 在 Plan 发出前校验 Constitution
7. Trace 文件系统写入，内容不可变
8. Constitution Enforcement Model 的 Pre-execution Guard 可用

**Non-Goals:**

- 不涉及浏览器自动化（Phase 2）
- 不涉及 Agent / LLM（Phase 3）
- 不涉及 MCP Server（Phase 4）
- CI 集成仅做基础 validation，不做 semantic replay（Phase 2）
- 不修改现有 Go 后端分层

## Decisions

**D1: Runtime 用 Go 实现**

虽然 Capability DSL 是语言无关的 YAML，但 Knowledge Runtime 是与现有 Go 后端（Gin + GORM）共存的系统组件。CLI Adapter 需要调用后端 API 执行 Plan，因此 Runtime 用 Go 实现，直接调用 `internal/domain/` 的 Repository 接口。

**D2: Trace 存储使用文件系统 + JSON**

Phase 1 Trace 不需要查询引擎。`.traces/YYYY/MM/DD/trace_<hash>.json` 的目录结构足够。JSON 序列化保证跨语言可读性（未来 Agent/MCP/CI 可直接消费）。Phase 3+ 可按需迁移到 Event Store。

**D3: Execution Plan plan_hash 使用 SHA256**

`plan_hash = SHA256(capability_version + resolved_rules_hash + inputs_schema + timestamp_ns)`。保证同一 capability 在不同时间点的 plan 有不同 hash，防止 replay 时的 hash 碰撞。

**D4: Pre-execution Guard 作为 Plan 生成管线中的中间件**

```
Parser → Resolver → Planner → Guard → Plan emitted
```

Guard 不是独立服务，而是 Runtime 中 `Plan()` 方法的最后校验步骤。失败时返回 `ErrConstitutionViolation`，不写入 Trace。

**D5: Capability DSL 首批定义覆盖核心业务流程**

Phase 1 定义 7 个 Capability：

1. `login` — 认证登录
2. `collect-rent` — 收租金（最核心高频操作）
3. `create-contract` — 创建合同
4. `issue-receipt` — 生成收据
5. `sign-contract` workflow — 签合同流程（Composite）
6. `backup-database` — 数据库备份
7. `create-user` — 创建用户

覆盖系统所有业务入口：认证、签合同、收款、催缴、管理操作。

**D6: CLI 不实现完整 Plan 执行，只实现 Plan 到 API 调用映射**

CLI Adapter 的职责是验证 Knowledge Runtime 生成 Plan 的能力，以及在开发调试时可视化 Plan 和 Trace。Phase 1 不通过 CLI 实现完整的"执行 Execution Plan"——那是 Browser/MCP Adapter 的职责。CLI 通过调用现有 Go 后端 API 完成执行。

**D7: Constitution Conflict Resolution 不实现，只留接口**

Phase 1 的 Pre-execution Guard 仅支持单 § 检查。多 § 冲突时按 §0 Precedence 处理（触发最高优先级 § 的 Guard）。Conflict Resolution Model 作为 Phase 3 的已知缺口记录在案。

## 模块边界

```
runtime/
    cmd/
        kr/                 ← CLI 入口 (main.go)
    internal/
        config/             ← CLI 配置加载
        loader/             ← YAML 文件加载 + 解析
        model/              ← Go struct: Capability, Workflow, Rule, ExecutionPlan, Trace
        resolver/           ← 引用解析 (BR → Rule, workflow → capability DAG)
        planner/            ← Workflow 展开 + Execution Plan 生成
        guard/              ← Pre-execution Constitution Guard
        cache/              ← 内存缓存 (Capability/Rule/Workflow/Plan)
        snapshot/           ← Execution Plan 版本快照与 plan_hash 生成
        trace/              ← Trace 写入 (文件系统)
        explain/            ← kr explain 实现
```

## 数据流

```
knowledge/capabilities/collect-rent.yaml
    ↓ (loader.Parse)
Capability struct
    ↓ (resolver.Resolve)
ResolvedCapability (BR → Rule objects, deps → DAG)
    ↓ (planner.Plan)
ExecutionPlan (plan_hash, steps, rules, observability, permissions)
    ↓ (guard.Check)
Constitution validation (Pre-execution Guard)
    ↓ (passed: emit)
Plan emitted to Adapter (CLI)
    ↓ (Adapter execute)
调用后端 API → 完成业务操作 → Trace 写入文件系统
```

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| Capability DSL 定义初期不稳定，频繁改动 YAML 格式 | loader/ 设计为松耦合 parser + validator，格式变化只需修改 parser |
| Execution Plan 版本与现有后端 API 不匹配 | CLI 直接调用后端 API，不引入中间层 |
| Trace 文件系统存储无归档策略 | Phase 1 依赖 git 自动管理；长期按日期目录自动轮转 |
| Pre-execution Guard 在 Constitution 未稳定时阻碍开发 | Guard 支持宽松模式（`--validate=false`），开发阶段可跳过 |

## 参考文献

- **brainstorm-spec.md** — 完整架构设计（580 行，9 层架构 + Timeline）
- **system/constitution.md** — TCA Constitution v1.1（13 条全局公理）
- **proposal.md** — 变更提案（Why / What / Capabilities / Impact）
