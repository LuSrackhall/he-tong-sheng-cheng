## 1. Capability DSL（首批定义）

- [x] 1.1 创建 `knowledge/` 顶层目录及 `capabilities/`、`workflows/`、`rules/` 子目录
- [x] 1.2 编写 `knowledge/capabilities/login.yaml`（认证登录）
- [x] 1.3 编写 `knowledge/capabilities/collect-rent.yaml`（收租金）
- [x] 1.4 编写 `knowledge/capabilities/create-contract.yaml`（创建合同）
- [x] 1.5 编写 `knowledge/capabilities/issue-receipt.yaml`（生成收据）
- [x] 1.6 编写 `knowledge/capabilities/backup-database.yaml`（数据库备份）
- [x] 1.7 编写 `knowledge/capabilities/create-user.yaml`（创建用户）
- [x] 1.8 编写 `knowledge/capabilities/ensure-contract-active.yaml`（依赖能力）
- [x] 1.9 编写 `knowledge/rules/` BR-001 ~ BR-010 业务规则
- [x] 1.10 编写 `knowledge/workflows/sign-new-contract.yaml`（签合同流程）
- [x] 1.11 编写 `knowledge/workflows/renew-contract.yaml`（续签流程）

## 2. Knowledge Runtime（Go 实现）

- [x] 2.1 创建 `runtime/` 目录及 `internal/model/`（Capability/Workflow/Rule/ExecutionPlan/Trace 结构体定义）
- [x] 2.2 实现 `runtime/internal/cache/` 内存缓存（Capability/Rule/Workflow/Plan 四类）
- [x] 2.3 实现 `runtime/internal/loader/` YAML 加载器（递归扫描、解析、验证）
- [x] 2.4 实现 `runtime/internal/resolver/` 引用解析（BR → Rule、workflow → capability DAG）
- [x] 2.5 实现 `runtime/internal/planner/` Execution Plan 生成器（含 plan_hash 分配）
- [x] 2.6 实现 `runtime/internal/planner/` Workflow 展开与条件分支处理
- [x] 2.7 实现 `runtime/internal/snapshot/` Execution Plan 版本锁定
- [x] 2.8 实现 Runtime 启动完整性校验（引用完整、DAG 无环）
- [x] 2.9 编写 `runtime/` 单元测试（loader、resolver、planner 核心逻辑）

## 3. Pre-execution Constitution Guard

- [ ] 3.1 实现 `runtime/internal/guard/` Pre-execution Guard 框架（§-aware 检查接口）
- [ ] 3.2 实现 §1 Knowledge Authority 检查（capability_id 注册验证）
- [ ] 3.3 实现 §2 Capability Atomicity 检查（所有 step 可回溯）
- [ ] 3.4 实现 §3 Plan Immutability 检查（plan_hash 在 emit 前锁定）
- [ ] 3.5 实现 §4 Separation of Concerns 检查（Plan 不含业务逻辑指令）
- [ ] 3.6 实现 §9 UI as Derived 检查（Plan 不含 UI selector 信息）
- [ ] 3.7 实现 Guard 宽松模式（`--validate=false` 跳过）
- [ ] 3.8 编写 Guard 单元测试（合法 Plan、非法 Plan、边界情况）

## 4. CLI Adapter

- [ ] 4.1 创建 `cli/kr/main.go` 入口（cobra 或标准 flag 解析）
- [ ] 4.2 实现 `kr plan <capability>` 命令（预览 Execution Plan）
- [ ] 4.3 实现 `kr plan <workflow>` 命令（预览多步骤 Plan）
- [ ] 4.4 实现 `kr run <capability> --inputs...` 命令（生成 Plan → Guard → 调用后端 API）
- [ ] 4.5 实现 `kr explain --trace <id>` 命令（读取并展示 Trace）

## 5. Execution Trace

- [ ] 5.1 实现 Trace 数据结构（identity、context、steps、observability、determinism、summary）
- [ ] 5.2 实现 Trace 文件写入器（`.traces/YYYY/MM/DD/trace_<plan_hash>.json`）
- [ ] 5.3 实现 Trace 读取器（按 trace_id 读取）
- [ ] 5.4 实现 CLI Adapter 的 Trace 写入集成（runde 后自动写入）
- [ ] 5.5 实现 CLI Adapter 的 determinism score（CLI：0.99）

## 6. CI 基础验证（Phase 1 最小版本）

- [ ] 6.1 实现 CI Knowledge Runtime validation 脚本（引用完整性、DAG 无环）
- [ ] 6.2 将 CI validation 接入 `.github/workflows/ci.yml`
- [ ] 6.3 编写集成测试：`kr plan` 覆盖所有 Capability、`kr run` 覆盖核心流程

---

## Post-Implementation Workflow

<!-- DO NOT MODIFY THIS SECTION — it defines the required workflow after all tasks are complete -->

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run `/opsx:verify` to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, go to main branch and merge (must ask user)
4. **Archive**: Run `/opsx:archive` on main
5. **Cleanup**: `git worktree remove .worktrees/change/<name>`

**Iteration**: If user does not accept, analyze the issue and recommend:
fix in place / new change / git reset + stash / git reset / abandon.
