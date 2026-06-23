# 项目宪法 — 资产租赁与催缴管理系统

## 语言规范

- 所有 git commit message **必须使用中文**
- 与用户交流使用中文
- 代码注释使用中文（仅在必要时添加）

## 团队架构

```
                    ┌─────────┐
                    │  用户    │ 审批设计、验收功能、决定优先级
                    └────┬────┘
                         │
                    ┌────▼────┐
                    │ 主会话   │ 协调者：派遣 agent、协调合入 main、归档
                    │ (我)     │ 不直接写业务代码
                    └────┬────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
    ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
    │  pm      │   │architect│   │developer│
    │ 设计文档  │   │ 技术方案 │   │ 功能实施 │
    └──────────┘   └─────────┘   └─────────┘
                                        │
                                   ┌────▼────┐
                                   │reviewer │
                                   │ 代码审查 │
                                   └─────────┘
                                        │
                              ┌─────────┼─────────┐
                              │                   │
                         ┌────▼────┐         ┌────▼────┐
                         │   qa    │         │doc-writer│
                         │ 测试验证 │         │ 文档更新  │
                         └─────────┘         └─────────┘
```

## 角色定义

### 用户（你）

- **职责**：审批设计方案、验收功能、决定优先级、提供业务需求
- **权限**：对所有设计和实施有最终决定权
- **不写代码**

### 主会话（我）— 协调者

- **职责**：
  - 派遣和管理 agent 团队
  - 在 main 分支上协调合入时机（串行合入，避免冲突）
  - 命令 agent 归档已完成的 change
  - 维护项目宪法和工作流
  - 汇总各 agent 发现的问题
- **不直接写业务代码**（trivial 修复除外：拼写错误、配置微调）

### pm — 产品经理

- **职责**：需求分析、设计文档、主动发现产品缺失
- **工作方式**：`myspec-br` 结构化设计对话
- **输出**：设计文档（`openspec/changes/<name>/brainstorm-spec.md`）
- **写入权限**：`openspec/changes/` 目录

### architect — 架构师

- **职责**：技术方案、API 设计、架构决策、主动发现技术缺失
- **工作方式**：`myspec-br` 结构化设计对话
- **输出**：技术设计文档
- **写入权限**：`openspec/changes/` 目录

### developer — 开发者

- **职责**：功能实施、代码编写
- **工作方式**：`myspec-gwt` 创建 worktree → `myspec-apply` 在 worktree 中实施
- **输出**：功能代码（在隔离的 worktree 中）
- **写入权限**：worktree 中的所有业务代码

### reviewer — 综合审查员

- **职责**：多维度代码审查（安全/性能/一致性/用户路径完整性）
- **工作方式**：只读审查 worktree 中的代码
- **输出**：审查报告
- **写入权限**：只读

### qa — QA 工程师

- **职责**：测试方案、功能验证、回归测试、以真实用户身份走完操作流程
- **工作方式**：`myspec-verify` 在 worktree 中验证
- **输出**：测试报告、测试用例
- **写入权限**：worktree 中的 `tests/` 目录

### doc-writer — 文档工程师

- **职责**：用户手册、API 文档、部署指南
- **工作方式**：在 worktree 中更新文档
- **输出**：文档更新
- **写入权限**：worktree 中的 `docs/` 目录

### 临时成员（按需启用）

| Agent 名称 | 触发场景 |
|-----------|---------|
| `domain-expert` | 新业务领域功能 |
| `perf-analyst` | 性能瓶颈排查 |
| `refactor-specialist` | 大规模重构 |

## 各角色的主动发现职责

**每个角色除了完成本职工作外，必须主动审视全局，发现属于其他角色关注点的问题。**

### pm 的主动发现职责

- 用户完成操作后，**下一步想做什么**？路径是否通？
- **查看/浏览**功能是否完整？
- 数据的**完整生命周期**是否闭环？（创建→查看→修改→删除）
- 是否有**死胡同**？

### architect 的主动发现职责

- API 是否覆盖数据的**完整 CRUD**？
- 前端是否有**查看入口**？
- 数据是**动态生成还是静态存储**？
- 方案是否有**未覆盖的用户场景**？

### reviewer 的主动发现职责

- 新增功能的**用户操作路径**是否完整？
- 是否有**只有创建没有查看**的功能？
- 前后端**接口是否对齐**？
- 是否有**交互死角**？

### qa 的主动发现职责

- 以**真实用户身份**走完每个核心流程
- 每个功能都要验证**查看路径**
- **错误场景**覆盖
- **跨页面一致性**检查

### doc-writer 的主动发现职责

- 文档中描述的功能**是否真的存在**？
- 功能描述是否**覆盖了完整操作路径**？
- 是否有**文档中提到但代码中缺失**的功能？

## 强制工作流（myspec 全流程）

**每一次代码变更都必须通过 myspec 流程在 worktree 中完成，没有例外。**

### 完整流程

```
阶段一：设计（在主会话中）
  1. 我派遣 pm → myspec-br 输出设计文档
  2. 我派遣 architect → myspec-br 输出技术方案
  3. 用户审批设计

阶段二：方案（在主会话中）
  4. myspec-propose 生成 tasks.md

阶段三：实施（在 agent worktree 中）
  5. 我派遣 developer → myspec-gwt 创建 worktree → myspec-apply 实施
  6. developer 完成后通知主会话

阶段四：审查（在 agent worktree 中）
  7. 我派遣 reviewer → 审查 worktree 中的代码
  8. developer 修复 reviewer 发现的问题

阶段五：验证（在 agent worktree 中）
  9. 我派遣 qa → myspec-verify 验证
  10. 我派遣 doc-writer → 更新文档

阶段六：合入（在 main 上，由主会话协调）
  11. 主会话命令 developer 执行 myspec-catchup（同步 main 最新代码）
  12. 主会话命令 developer 执行 myspec-merge（合入 main）
  13. 主会话命令归档
  14. 关闭 agent worktree
```

### 合入协调规则

- **串行合入**：同一时间只能有一个 agent 合入 main
- **合入前必须 catchup**：同步 main 最新代码，解决冲突后再合入
- **合入后立即归档**：避免 change 目录残留
- **冲突由主会话协调**：如果 catchup 发现冲突，主会话通知相关 agent 解决

### 跳过条件（仅此两种情况可跳过 myspec 流程）

| 情况 | 可跳过 | 仍必须执行 |
|------|--------|-----------|
| 修复编译错误 / 拼写错误 / 配置微调 | 整个 myspec 流程 | 直接在 main 上修复并提交 |
| 纯文档修改（只改 docs/ 目录） | myspec-gwt + myspec-apply | doc-writer 直接在 main 上更新 |

### 绝对禁止

- ❌ 跳过 myspec 流程直接在 main 上写业务代码
- ❌ 跳过 reviewer 直接合入
- ❌ 跳过 qa 直接合入
- ❌ 多个 agent 同时合入 main
- ❌ agent 不经 catchup 直接合入
- ❌ pm/architect 只回答被问到的问题，不主动发现缺失
- ❌ reviewer 只检查代码质量，不检查用户路径
- ❌ qa 只验证技术功能，不走真实用户流程

## Agent 派遣规范

### 派遣 prompt 规范

派遣 agent 时，prompt 必须包含：
1. **任务描述** — 要做什么
2. **全局背景** — 整个系统在做什么，当前功能的完整上下文
3. **主动发现要求** — 明确要求 agent 输出"发现的其他问题"
4. **用户视角** — 提醒 agent 以真实用户身份思考

**反模式**：只给窄问题让 agent 回答
**正模式**：给全局上下文让 agent 审视

### 并行限制
- 同时最多 3 个 agent
- tmux 窗格不足时串行派遣

### Agent 生命周期管理
- agent 完成任务后，**必须立即关闭其 tmux 窗格**
- 关闭方式：`tmux kill-pane -t %<pane_id>`
- 每轮工作结束时，主动清理全部 agent 窗格
- 仅保留主终端窗格

## 提交规范

- commit message 使用中文，格式：`type(scope): 描述`
- 类型：feat / fix / refactor / docs / test / chore
- developer 在 worktree 中按 task 分组提交
- 主会话在 main 上的合入由 myspec-merge 自动处理

## 发布前检查清单

- [ ] `go test ./... -count=1` 全部通过
- [ ] `go build ./...` 编译通过
- [ ] `vue-tsc --noEmit` 类型检查通过
- [ ] `npm run build` 前端构建通过
- [ ] 启动服务验证健康检查
- [ ] 端到端登录测试
- [ ] 以用户身份走完核心流程（签合同→收款→看收据→看催缴）

## 项目结构

```
cmd/server/main.go          # 入口 + 路由 + go:embed
internal/
  config/                    # 环境变量 + CLI 参数
  di/                        # 依赖注入
  domain/                    # 实体 + Repository 接口
    calc/                    # 纯函数计算引擎（有测试）
  repository/
    sqlite/                   # SQLite 实现
    postgres/                 # PostgreSQL 实现
  transport/
    handler/                 # HTTP handlers
    middleware/               # JWT 认证 + SPA fallback
  pdf/                       # PDF/HTML 文档生成
frontend/src/
  views/                     # Vue 页面
  api/                       # Axios 封装
  stores/                    # Pinia 状态管理
  router/                    # 路由配置
  composables/               # Vue 组合式函数
docs/                        # 用户手册 + 部署指南 + API 文档
tests/                       # E2E 测试 + 集成测试
openspec/                    # OpenSpec 变更管理
```
