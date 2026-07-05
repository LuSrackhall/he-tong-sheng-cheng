# TCA Constitution v1.1

> Task-Centric Architecture 的不可违反语义约束层。
>
> 本文件是系统级全局不变式（global invariant），不是 Change artifact。
> 只能 append，不能修改旧条。版本升级时新旧版本并存。

---

## §0 Precedence of Truth Sources

系统所有语义判断遵循以下优先级（从高到低）：

1. **Execution Trace** — 事实优先（发生过什么）
2. **Execution Plan** — 契约优先（应该发生什么）
3. **Capability / Rule** — 语义定义（能力是什么）
4. **CI Analysis** — 统计判断（趋势如何）
5. **Adapter Feedback** — 执行观察（执行者看到了什么）
6. **Human Annotation** — 解释性信息（人工说明）

**任何低优先级来源不得覆盖高优先级来源的语义结论。**

---

## 第一簇：Authority（权威来源）

### §1 Knowledge Authority

Knowledge Layer is the sole source of truth for capability definition. No other layer defines what the system can do.

### §2 Capability Atomicity

All execution must originate from a Capability or Workflow definition. Ad-hoc execution paths outside the Knowledge Layer are forbidden.

### §3 Plan Immutability

Execution Plan is immutable once emitted by Knowledge Runtime and assigned a `plan_hash`. No layer may modify a frozen plan.

---

## 第二簇：Boundary（层边界）

### §4 Separation of Concerns

Three invariant layers: Runtime generates Plans, Adapters execute Plans, Domain systems execute business logic. No cross-layer role substitution.

### §5 Adapter Purity

Adapters must not perform reasoning, planning, or business interpretation. They only execute, observe, and report.

### §6 No Semantic Backflow

Trace and Feedback cannot modify Execution Plan directly. Only Knowledge Layer may evolve system semantics.

---

## 第三簇：Truth（不可变事实）

### §7 Trace Irreversibility

Execution Trace is immutable. It represents historical truth that cannot be altered, replayed differently, or deleted.

### §8 CI as Observer

CI must not alter system behavior. It observes, compares, and reports — with no hidden remediation or self-healing logic.

### §9 UI as Derived Artifact

UI mappings are derived from Execution Plan, not authoritative. Knowledge Layer always governs UI semantics, not the reverse.

---

## 第四簇：Evolution（演化规则）

### §10 Determinism Gradient

Execution modes follow a strict determinism hierarchy: CLI > Browser > Exploration. A lower-determinism result cannot override a higher-determinism result for the same Capability.

### §11 Drift is Temporal

System correctness is evaluated over time-series traces, not single executions. CI gates on sustained drift trends, not isolated spikes.

### §12 Knowledge Evolution Only

Only the Knowledge Layer may evolve system behavior. Runtime, Adapters, Traces, and CI are invariant layers with no self-evolution authority.

---

## 附录 A：Constitution 版本与修改记录

| 版本 | 日期 | 修改内容 |
|------|------|----------|
| v1.0 | 2026-07-04 | 初始 12 条公理（Authority / Boundary / Truth / Evolution） |
| v1.1 | 2026-07-04 | 新增 §0 Precedence of Truth Sources |
| v1.2 | 2026-07-05 | Phase 2 扩展：§5 Adapter Purity、§7 Trace Irreversibility、§10 Determinism Gradient 加入 Runtime Guard。§6、§8、§11、§12 仍为 deferred。新增 Playwright E2E 测试套件。 |

**修改规则：** Constitution 只能 append，不能修改旧条。版本升级时新旧版本并存。Runtime 始终只绑定一个 active version。
