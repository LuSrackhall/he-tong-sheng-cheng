# Task-Centric Architecture（TCA）设计文档

## Context

资产租赁与催缴管理系统已经具备成熟的生产能力。业务模型围绕"合同"锚点、SQLite/PostgreSQL 双数据库、`go:embed` 单一二进制部署、JWT 认证等均已就绪。但存在两个核心痛点：

1. **缺乏 E2E 测试** — 核心业务流程（签合同、收租金、催缴）没有自动化验证
2. **缺乏用户体验感受** — 系统虽然功能正确，但无人从"第一次使用"角度评估

这两个痛点实际上是同一问题的两面：**缺乏对"业务任务"的形式化定义**。

本设计从 E2E 测试需求出发，逐步演进为一套完整的 **Task-Centric Architecture（TCA）**——以业务能力（Capability）为中心、Execution Plan 为纽带、多运行环境适配器为终端的执行语义操作系统。

## Goals / Non-Goals

### Goals

1. 定义以 **Capability** 为核心的一等公民 DSL，描述系统提供的所有业务能力
2. 构建 **Knowledge Runtime**，负责加载、解析、解析引用、规划执行、查询 Capability
3. 设计 **Execution Plan** 作为运行时唯一契约，不依赖任何 UI/API 细节
4. 实现 **Runtime Adapters**（CLI、Browser、Agent、MCP），将同一 Plan 映射到不同执行环境
5. 建立 **Execution Trace** 作为不可变系统记忆，支持回放、可解释、可学习
6. 构建 **CI 系统**作为 Execution Graph 变化影响分析器，而非传统测试运行器
7. 形成 **Feedback Loop**，所有系统优化来源于 Trace，驱动 Knowledge/UI/Code 演进

### Non-Goals

- 不引入新的 UI 框架（保持自定义 CSS）
- 不重构现有 Go 后端分层（domain / repository / transport）
- 不新增业务功能（不改变现有领域模型）
- 不替换 OpenSpec 工作流（TCA 作为 OpenSpec 之上的一层知识层）
- 不在 Phase 1 涉及浏览器自动化或 Agent

## Decisions

### D1: Capability DSL 取代传统 Task 定义

**选择：** Task 不应描述"一次调用"，而应描述"一个业务能力"。格式为 YAML，存放于顶层 `knowledge/` 目录。

**理由：**
- Capability 是所有执行环境的共同语义锚点
- YAML 语言无关，Go/JS/Python 均可消费
- 不放在 `internal/domain/`（Go 私有），也不放在 `openspec/`（被视为 spec），而是放在顶层 `knowledge/`

**关键格式：**

```yaml
# knowledge/capabilities/collect-rent.yaml
id: collect-rent
version: 1
status: stable
since: 1.3.0

metadata:
  tags: [rent, payment, receipt]
  owner: payment-team
  priority: P0

title: 收租金
goal: 录入一笔租金收款并生成收据

inputs:
  - { name: contract, domain: Contract }
  - { name: amount,   domain: Money, min: 0.01 }
  - { name: date,     domain: Date }
  - { name: notes,    domain: String, required: false }

outputs:
  - { name: payment, domain: Payment }
  - { name: receipt, domain: Receipt }

preconditions:
  - BR-001  # Contract must be active or arrears
  - BR-002  # Payment amount > 0
  - BR-003  # ReceiptBook available

effects:
  sync:
    creates:
      - Payment
      - Receipt
    updates:
      - Contract.TotalReceived
  async:
    emits:
      - PaymentCreated

permissions:
  - rent:collect

dependencies:
  requires:
    - ensure-contract-active
  triggers:
    - refresh-arrears
    - update-dashboard

observability:
  success:
    - "Payment 记录存在，金额匹配"
    - "Receipt 已生成，编号有效"
    - "Contract.TotalReceived 已增加"
  failure:
    - "BR-004 违反（超额支付）"
    - "BR-003 违反（收据本耗尽）"
```

**Rule 独立存放：**

```yaml
# knowledge/rules/BR-001.yaml
id: BR-001
title: 合同状态要求
severity: error
description: 合同状态必须为 active 或 arrears
applies_to:
  - collect-rent
  - issue-receipt
  - renew-contract
```

**Workflow 支持条件分支：**

```yaml
# knowledge/workflows/sign-new-contract.yaml
id: sign-new-contract
steps:
  - if: asset.exists
    then: search-asset
    else: create-asset
  - if: tenant.exists
    then: search-tenant
    else: create-tenant
  - create-contract
  - issue-receipt
```

**关键原则：**
- 不描述 UI、API、页面、按钮 — 0 个实现细节
- 不写 Actor — 谁都可以执行，权限控制
- effects 分 sync/async — side effect 不污染成功标准
- Business Rules 编号化 — 可被 Playwright/Agent/MCP 引用
- Dependencies 构成可推导的 DAG
- 版本化 — 支持业务规则演进

### D2: Knowledge Runtime 是内核，不是 Loader

**选择：** Knowledge Engine 不只是 YAML 加载器，而是五层运行时：

```
Storage → Parser → Resolver → Planner → Query
```

**理由：**
- Executor 不应该消费原始 YAML，而应消费 Execution Plan
- Planner 负责展开 Workflow 为 Capability DAG（处理条件分支）
- Resolver 负责引用解析（BR → Rule、workflow → capability）
- Cache 支持 Capability/Rule/Workflow/Execution Plan 缓存

**Engine 接口：**

```go
type Runtime interface {
    Plan(goal string) (*ExecutionPlan, error)
    LoadCapability(id string, version ...string) (*Capability, error)
    LoadRule(id string) (*Rule, error)
    ResolveWorkflow(id string) (*WorkflowPlan, error)
    FindCapabilities(filter Filter) ([]*Capability, error)
    Explain(plan *ExecutionPlan, trace *Trace) (*Explanation, error)
}
```

**系统边界（关键）：** Knowledge Runtime 不执行业务逻辑。它只产生 Execution Plan。实际业务逻辑由领域层（Go backend `internal/domain/`）执行。Runtime ≠ workflow engine ≠ application server。

```
Runtime     → 生成 Plan
Adapter     → 执行 Plan
Domain (Go) → 执行业务逻辑
```

### D3: Execution Plan 是唯一契约

**选择：** Knowledge Runtime 和 Adapter 之间只通过 Execution Plan 通信。

**生命周期约束：**
- Execution Plan 在 Runtime 发出并分配 `plan_hash` 之后**不可变**
- Execution Plan is frozen at the moment of Trace creation
- Adapter 不可回写或修改 Execution Plan
- Adapter 只能产生 Execution Trace

**内容：**
```
ExecutionPlan
  ├── identity { capability_id, version, plan_hash }
  ├── resolved_rules [{ rule_id, description }]
  ├── resolved_dependencies DAG
  ├── input_schema, output_schema
  ├── permissions
  ├── execution_profile (strict | exploratory | production | debug)
  ├── ui_mapping_version
  └── observability { success, failure, checks }
```

### D4: Runtime Adapters 只消费 Execution Plan

**选择：** Adapter 是 Knowledge Runtime 的终端，不负责推理。

```go
type Adapter interface {
    Name() string
    Capabilities() []CapabilityDescriptor
    CanHandle(capabilityID string) bool
    Execute(ctx Context, plan ExecutionPlan) Result
    Explain(plan ExecutionPlan, result Result) Explanation
}
```

**Adapter 类型：**

| Adapter | Profile | 用途 | Phase |
|---------|---------|------|-------|
| CLI | strict | 调试、开发、验证 Knowledge Runtime | 1 |
| Browser | semantic | CI 回归、UI mapping drift 检测 | 2 |
| Exploration | exploratory | UX Audit、混沌测试、可发现性测试 | 3 |
| MCP | production | 生产环境 Agent Runtime | 4 |

**UI Action Plan（Browser Adapter 内部产物）：**
```
Execution Plan → UI Action Plan → Playwright
```
- UI Action Plan = 可缓存的推导结果，不是权威来源
- Capability 永远高于 UI
- Q 版本化：`adapters/browser/mapping/collect-rent/v1.yaml`

### D5: Execution Trace 是系统记忆

**选择：** 每次执行产生一条不可变的 Trace，作为系统唯一 truth source for behavioral learning。

```yaml
trace:
  identity: { trace_id, capability_id, adapter, plan_version, plan_hash }
  context: { user, session, timestamp, environment }
  execution_profile: strict | exploratory | production | debug
  plan_snapshot: (version locked)
  steps:
    - intent_step: "Fill Amount"
      runtime_step:
        selector: "#amount-input"
        action: "type"
      input: { value: 5000 }
      output: { success: true }
      duration_ms: 120
  observability:
    checks: [{ rule_id: BR-001, passed: true }, ...]
  determinism:
    score: 0.75
    factors:
      ui_stability: 0.8
      api_consistency: 0.9
      agent_variance: 0.0
  summary: { duration_ms, step_count, error_count }
```

**核心原则：**
- Trace ≠ 日志。Trace = Execution Plan 在真实世界中的完整状态快照
- Trace 不可逆：All execution is irreversible at the trace level. Only knowledge layer evolves.
- Trace 不能在任何情况下修改 Execution Plan
- Feedback may suggest changes, but only Knowledge Layer can modify Execution Plan
- 所有系统优化只能来源于 Trace，不能来源于人工经验
- `plan_hash` 是 Trace integrity 的关键字段，用于 CI diff / replay 校验 / drift detection

### D6: CI 是变化影响分析系统

**选择：** CI 不是测试运行器，而是 Execution Graph 的语义一致性验证器。

**核心能力：**
1. **Validation** — Knowledge Runtime 完整性检查（引用完整、DAG 无环）
2. **Replay** — CLI exact replay + Browser semantic replay
3. **Drift 分析** — 三维评分：structural / semantic / ui
4. **Gating** — strict（CLI）阻断 PR；browser（语义）警告不阻断

**关键约束：**
- CI 不验证 steps，CI 验证 semantic outcomes
- Drift is evaluated over a rolling window of traces, not a single execution
- Any capability with **degrading determinism trend over time** must trigger investigation
- CI does not block on drift spikes, but blocks on sustained drift degradation trend

**Exploration Adapter 降级为异步后处理：**
```
CI Pipeline → Trace Replay → Drift Analysis → Trigger Exploration (async)
```

### D7: Feedback 结构化分流

**选择：** 所有反馈使用统一 Schema，按类型分流到 Product / UI / Code / Capability。

```yaml
feedback:
  adapter: browser
  type: ui | code | capability | workflow
  severity: critical | high | medium | low | info
  capability: collect-rent
  scenario: happy-path
  title: 收款按钮不明显
  description: "..."
  recommendation: "将'收款'按钮从折叠菜单移出"
```

## 目录结构

```
knowledge/                          # 业务知识层（领域资产，不属于任何实现）
    capabilities/
        collect-rent.yaml           # 业务能力定义
        create-contract.yaml
        issue-receipt.yaml
        login.yaml
        ...
    workflows/
        sign-new-contract.yaml      # 复合流程编排
        renew-contract.yaml
        ...
    rules/
        BR-001.yaml                 # 业务规则（独立、可复用）
        BR-002.yaml
        ...
    glossary/                       # 领域术语
        contract.md
        arrears.md
        ...

runtime/                            # Knowledge Runtime 实现
    loader/                         # YAML 加载
    model/                          # Capability, Workflow, Rule, Plan 模型
    resolver/                       # 引用解析、DAG 构建
    planner/                        # Workflow 展开、条件分支
    cache/                          # Capability/Rule/Plan 缓存
    snapshot/                       # Execution Plan 版本快照

cli/                                # CLI Adapter (实际在 runtime/cmd/kr/ 实现)
    main.go                         # kr run / kr plan / kr explain

adapters/                           # Runtime Adapters
    browser/
        mapping/                    # Capability → UI Action Plan（版本化）
            collect-rent/v1.yaml
            ...
        playwright/                 # Playwright test 生成器
    exploration/
        personas/                   # Persona prompt 模板
        modes/                      # UX / Chaos / Regression / Accessibility
        reporter/                   # Feedback → Trace 关联
    mcp/                            # Phase 4

ci/                                 # CI 集成
    validate/                       # Runtime integrity check
    replay/                         # Semantic replay engine
    drift/                          # 三维 drift scoring

.traces/                            # Trace 存档（不可变）
    2026/07/04/trace_abc123.json

e2e/                                # Playwright 测试（Browser Adapter 产物）
    capabilities/
        collect-rent.spec.ts
        ...
    workflows/
        sign-new-contract.spec.ts
```

## 九层架构全景

```
                      Goal / Intent
                           │
                           ▼
 ┌──────────────────────────────────────────────────────┐
 │                 Knowledge Runtime                     │
 │  knowledge/ → Parser → Resolver → Planner → Query    │
 │  Output: Execution Plan (version locked, ui mapped)  │
 └──────────────────────────┬───────────────────────────┘
                            │
                            ▼
                    Execution Plan
              (the only contract between
               Runtime and Adapter)
                            │
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                  ▼
    ┌────────────┐   ┌────────────┐   ┌────────────┐
    │   CLI     │   │  Browser   │   │Exploration │
    │  Adapter  │   │  Adapter   │   │  Adapter   │
    │(strict)   │   │(semantic)  │   │(async)     │
    └─────┬─────┘   └─────┬──────┘   └──────┬─────┘
          │               │                 │
          └───────────────┴─────────────────┘
                          │
                          ▼
                   Execution Trace
             (immutable, versioned, scored)
                          │
                          ▼
              ┌──────────────────────┐
              │  CI System           │
              │  (change impact)     │
              └──────────┬───────────┘
                         │
              ┌──────────┴───────────┐
              │                      │
              ▼                      ▼
       Feedback Loop          Trace Archive
       (structured)           (learning data)
              │
              ▼
   Knowledge / UI / Code improvement
```

## System Definition

> **The Knowledge Runtime is the operating system for business capabilities.**
> **Knowledge Runtime does not execute business logic. It only produces Execution Plans.**
> **Execution Plan is the runtime-neutral contract.**
> **Runtime Adapters are the execution environments.**
> **Trace is the immutable memory.**
> **CI is the semantic consistency validator over execution graphs.**
> **Feedback is the evolution driver.**
> **All execution is irreversible at the trace level. Only knowledge layer evolves.**

## Timeline / Rollout Plan

### Milestone 1: Kernel（2-3 周）

**目标：** Knowledge Runtime 跑通，CLI Adapter 能调 capability

- `knowledge/capabilities/` 首批 5-8 个核心 capability
- `knowledge/rules/` BR-001 ~ BR-010
- `knowledge/workflows/` sign-contract, renew-contract
- `runtime/loader/` — YAML 加载
- `runtime/resolver/` — 引用解析
- `runtime/planner/` — workflow 展开（最小版本）
- `runtime/cache/` — 内存缓存
- `runtime/snapshot/` — Execution Plan 版本锁定
- CLI: `kr run`, `kr plan`, `kr explain`
- Phase 1 中 `runtime/internal/snapshot/` 预留但未接入执行管线
- Trace 文件正确写入 `.traces/`

**交付标准：** `kr run collect-rent --contract 1 --amount 500` 通过 CLI 成功

### Milestone 2: Reference Browser Adapter + CI（3-4 周）

**目标：** Capability → UI 映射建立，CI 能做 Semantic Replay

- `adapters/browser/mapping/` 首批 5 条 Capability 的 UI Action Plan
- `adapters/browser/playwright/` Playwright test 生成器
- `ci/validate/` — Knowledge Runtime 完整性检查
- `ci/replay/` — Semantic Replay 引擎
- `ci/drift/` — 三维 Drift Score: structural / semantic / ui
- CI pipeline: blocking（CLI strict）+ non-blocking（Browser semantic）
- Exploration Adapter 作为异步后处理接入

**优先 Capability：** login → collect-rent → create-contract → issue-receipt → sign-contract workflow → backup-database → create-user

### Milestone 3: Exploration Adapter（3-4 周）

**目标：** Agent 驱动的 UX 审计、混沌测试、可发现性测试

- `adapters/exploration/personas/` 4 种 Persona prompt 模板
- `adapters/exploration/modes/` UX / Chaos / Regression / Accessibility
- `adapters/exploration/reporter/` Feedback → Trace 关联
- `feedback/collector/` 结构化 Feedback 聚合
- `feedback/dashboard/` UX Score 趋势

**关键约束：** `exploration.budget = { max_steps: 200, max_time_ms: 300000, max_capabilities: 10 }`

### Milestone 4: MCP + Planner（4-6 周）

**目标：** Capability DSL 经过数月验证后，开放生产环境 Agent Runtime

- `adapters/mcp/tool-generator/` Capability → MCP tool definition
- `adapters/mcp/server/` MCP Server
- `runtime/planner-v2/` 智能 goal 拆解
- `runtime/executor-registry/` 自适应调度

**安全约束：** MCP must be Capability-sandboxed. Cannot bypass Knowledge Runtime.

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| Capability DSL 定义初期不稳定，频繁改动 | Milestone 1 的 CLI 允许快速迭代 DSL 而无需 UI/CI |
| Browser Adapter flaky（Playwright 不稳定） | 用 Semantic Replay 而非 Exact Replay 容忍 UI 抖动 |
| Agent Executor 输出不稳定 | 降为异步后处理，不阻断 CI |
| Execution Plan snapshot 版本管理复杂度 | Phase 1 只做文件级版本锁定，不引入数据库 |
| Trace 存储量随时间增长 | Phase 1 用文件系统自动轮转，Phase 3+ 可迁移到 Event Store |
| UI 频繁变化导致 UI mapping drift 常态化 | CI 不阻断单次 drift，但监控持续下降趋势触发告警 |

## Constitution Enforcement Model

Constitution 如果只是文档，就仍然是"设计原则"而非"系统层约束"。必须被 Runtime 可计算地执行（computationally enforceable）。

### 三层 Enforcement Gate

```
              Constitution (13 axioms)
                      │
        ┌─────────────┼─────────────┐
        ▼             ▼             ▼
  Pre-execution   Runtime      Post-execution
  Guard           Guard        Audit
  (Plan-time)     (Adapter)    (Trace)
```

#### 1. Pre-execution Guard（Plan-time check）

**触发点：** Knowledge Runtime 生成 Execution Plan 之后、发出之前

**检查项：**

| 宪法条 | 检查 | 实现方式 |
|--------|------|----------|
| §1 Knowledge Authority | Plan 的 capability_id 必须在 knowledge/ 中注册 | capability registry lookup |
| §2 Capability Atomicity | 所有 step 可回溯到 capability 或 workflow | DAG source tracking |
| §3 Plan Immutability | plan_hash 在 emit 前不可变 | hash lock |
| §4 Separation | Plan 不包含业务逻辑执行指令 | schema validation |
| §9 UI as Derived | Plan 不包含 UI selector 信息 | field validation |
| §12 Knowledge Evolution | Plan 不包含 self-modification 指令 | pattern detection |

**失败处理：** Plan rejected。Runtime 返回 `ErrConstitutionViolation{axiom, detail}`。不写入 Trace。

#### 2. Runtime Guard（Adapter check）

**触发点：** Adapter 执行 Execution Plan 期间

**检查项：**

| 宪法条 | 检查 | 实现方式 |
|--------|------|----------|
| §5 Adapter Purity | Adapter 不修改 Plan、不注入新 step | step hash chain |
| §6 No Semantic Backflow | Adapter 不可回写 Feedback 到 Plan | write permission boundary |
| §10 Determinism Gradient | Adapter 不隐式降级自己的 execution profile | profile check |

**失败处理：** 当前 step 标记为 `constitution_violation`，Adapter 可继续或中止（取决于 execution_profile：strict 必须中止，exploratory 可继续但记录）。

#### 3. Post-execution Audit（Trace check）

**触发点：** Trace 写入后、进入 CI 前

**检查项：**

| 宪法条 | 检查 | 实现方式 |
|--------|------|----------|
| §7 Trace Irreversibility | Trace 写入后 hash 不变 | content-addressable storage |
| §8 CI as Observer | CI 不修改 Trace | read-only CI pipeline |
| §10 Determinism Gradient | Trace 中 execution_profile 与实际行为一致 | profile attestation |
| §11 Drift is Temporal | drift 报告基于时间窗口而非单点 | windowed aggregation check |

### 已知缺口：Constitution Conflict Resolution

**问题：** 多个宪法条在同一执行上下文可能产生约束冲突。例如 Adapter 在 exploratory mode 下同时受 §5 (Purity) 和 §10 (Determinism Gradient) 约束，当探索行为被判定为"非纯执行"时，二者产生矛盾。

**Phase 1 处理原则（不阻断实现）：**

| 场景 | Phase 1 行为 | 未来方向 |
|------|-------------|----------|
| 单 § 违反 | Pre-execution Guard 拒绝 Plan | 保持不变 |
| 多 § 冲突 | 触发最低编号 § 的 Guard | Phase 3 引入 Conflict Resolution：约束合成 + 上下文仲裁 + 违反分类 |
| Guard 间不一致 | 以 Pre-execution Guard 为准（§0 最高优先级）| 保持不变 |

**关键约束：** Phase 1 的任何实现假设不得要求 Conflict Resolution 层在 Phase 3 之前不存在。所有 Guard 实现必须是可组合的（composable），不硬编码跨-§ 仲裁逻辑。

**状态：** 已知缺口，Phase 3 填补。

**失败处理：** 不阻断（Trace 已存在）。生成 `ConstitutionAuditReport` 作为 Feedback 输入。

### Enforcement 与 Phase 的关系

| Phase | 实现的 Gate | 理由 |
|-------|-------------|------|
| 1 (Kernel) | Pre-execution Guard | CLI Adapter 执行前校验 Plan 合法性 |
| 2 (Browser + CI) | Post-execution Audit | CI 引入后需要验证 Trace 与 Constitution 一致 |
| 3 (Exploration) | Runtime Guard（部分） | Agent 执行需要约束 Adapter Purity |
| 4 (MCP) | Runtime Guard（完整） | 生产环境需要全量运行时检查 |

### Guard 的 Constitution 依据

三层 Gate 各自的权威来源遵循 **§0 Precedence of Truth Sources**：

- Pre-execution Guard 的依据是 §1–§4、§9、§12（高优先级定义）
- Runtime Guard 的依据是 §5、§6、§10（中优先级约束）
- Post-execution Audit 的依据是 §7、§8、§10、§11（低优先级验证）

**Guard 之间不冲突。** 如果 Pre-execution Guard 通过了，Runtime Guard 就不应因同一 § 拒绝。如果发生，说明 Constitution 有歧义，触发 Constitution 版本升级（appendix only）。

## 参考文献

- **TCA Constitution v1.1** — `system/constitution.md`（全局不变式层，13 条不可违反公理）
