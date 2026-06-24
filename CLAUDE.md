# 项目宪法 — 资产租赁与催缴管理系统

## 语言规范

- 所有 git commit message **必须使用中文**
- 与用户交流使用中文
- 代码注释使用中文（仅在必要时添加）

## 核心原则

1. **主会话 = 纯调度者**。不参与设计决策、不写业务代码、不进入 myspec 流程内部
2. **每个 change 的完整生命周期由一个 myspec agent 执行**。通过 Skill 工具调用 myspec 技能链（br → propose → apply → verify → merge → archive），全程在同一 agent 中完成
3. **每个 myspec agent 有专属答复团队**。由主会话预派遣，与 myspec agent 直接交互，主会话不充当中间人
4. **上下文隔离**。不同 change 的 myspec agent 和答复团队互相独立，防止跨 change 上下文污染
5. **主会话仅在合并时机上介入**。协调多个 change 对 main 的串行合入

## 团队架构

```
┌─────────┐
│  用户    │ 发起问题、最终验收
└────┬────┘
     │
┌────▼──────────────────────────────────────────┐
│  主会话（纯调度者）                              │
│  - 接收问题，派遣 myspec agent + 专属答复团队    │
│  - 管理多组 change 的并行状态                    │
│  - 通过 SendMessage 调度合并时机（串行合入 main）│
│  - 汇总状态报告给用户                           │
│  - 不进入设计/审查/验证决策                      │
└────┬──────────────────────────────────────────┘
     │ 预派遣（每组 change 一个 myspec agent + 一套答复团队）
     │
     ├── Change A ──────────────────────────────┐
     │   ┌─────────────────────────────────┐    │
     │   │ myspec agent A                  │    │
     │   │ Skill: br → propose → apply →   │    │
     │   │ verify → merge → archive        │    │
     │   └──────────┬──────────────────────┘    │
     │              │ 直接派遣/交互               │
     │   ┌──────────▼──────────────────────┐    │
     │   │ 答复团队 A（专属，防上下文污染）  │    │
     │   │ pm · architect · reviewer · qa   │    │
     │   └─────────────────────────────────┘    │
     │                                          │
     ├── Change B ──────────────────────────────┐
     │   ┌─────────────────────────────────┐    │
     │   │ myspec agent B                  │    │
     │   │ Skill: br → propose → apply →   │    │
     │   │ verify → merge → archive        │    │
     │   └──────────┬──────────────────────┘    │
     │              │ 直接派遣/交互               │
     │   ┌──────────▼──────────────────────┐    │
     │   │ 答复团队 B（专属，防上下文污染）  │    │
     │   │ pm · architect · reviewer · qa   │    │
     │   └─────────────────────────────────┘    │
     │                                          │
     └── ...（最多 3 组并行）                    │
```

## 角色定义

### 用户（你）

- **职责**：发起问题、最终验收功能是否解决问题
- **权限**：对最终验收有决定权
- **不写代码，不参与中间决策**

### 主会话（我）— 纯调度者

- **职责**：
  - 接收用户问题，派遣 myspec agent
  - 管理多个 change 的并行运行状态
  - 调度合并时机：当 myspec agent 报告"就绪待合入"时，按序安排合入 main
  - 汇总各 change 的进度报告给用户
- **不进入**：设计对话、代码审查、验证决策、工件生成
- **不写业务代码**（trivial 修复除外：拼写错误、配置微调）

### myspec agent — change 生命周期负责人

每个 change 派遣一个 myspec agent，负责从设计到归档的完整生命周期。

- **职责**：
  - 通过 Skill 工具依次调用 myspec 技能链：
    - `myspec-gwt` → 创建 worktree
    - `myspec-br` → 苏格拉底式设计对话
    - `myspec-propose` → 生成方案工件
    - `myspec-apply` → 按 task group 实施代码
    - `myspec-verify` → 验证实现
    - `myspec-merge` → catchup + 合入 + 归档
  - 在各技能执行过程中，按需派遣答复团队子智能体进行评估/决策/复核
  - 合并前通知主会话等待调度
- **权限**：worktree 中的所有文件读写
- **必须通过 Skill 工具调用技能，不得跳过或自行替代**

### 答复团队 — 专属子智能体（由主会话预派遣，与 myspec agent 直接交互）

**每个 myspec agent 拥有专属的答复团队实例，不同 change 的答复团队互不共享，防止上下文污染。**

答复团队是 myspec agent 的"参谋部"。myspec agent 在各技能执行过程中按需向他们提出评估请求，他们完成后将结论直接返回给 myspec agent。

#### pm（产品经理）

- **派遣时机**：设计阶段的需求澄清、产品方向决策
- **职责**：需求分析、主动发现产品缺失、评估用户路径完整性
- **主动发现**：
  - 用户完成操作后，**下一步想做什么**？路径是否通？
  - 数据的**完整生命周期**是否闭环？（创建→查看→修改→删除）
  - 是否有**死胡同**？

#### architect（架构师）

- **派遣时机**：设计阶段的技术方案评估、API 设计决策
- **职责**：技术方案、架构决策、主动发现技术缺失
- **主动发现**：
  - API 是否覆盖数据的**完整 CRUD**？
  - 前端是否有**查看入口**？
  - 方案是否有**未覆盖的用户场景**？

#### reviewer（审查员）

- **派遣时机**：实施完成后的代码审查
- **职责**：多维度代码审查（安全/性能/一致性/用户路径完整性）
- **主动发现**：
  - 新增功能的**用户操作路径**是否完整？
  - 是否有**只有创建没有查看**的功能？
  - 前后端**接口是否对齐**？

#### qa（QA 工程师）

- **派遣时机**：修复审查问题后的功能验证
- **职责**：以真实用户身份走完操作流程、验证查看路径、错误场景覆盖
- **主动发现**：
  - 每个功能都要验证**查看路径**
  - **错误场景**覆盖
  - **跨页面一致性**检查

#### doc-writer（文档工程师）

- **派遣时机**：验证通过后更新文档
- **职责**：用户手册、API 文档、部署指南

#### 临时成员（按需启用）

| Agent 名称 | 触发场景 |
|-----------|---------|
| `domain-expert` | 新业务领域功能 |
| `perf-analyst` | 性能瓶颈排查 |
| `refactor-specialist` | 大规模重构 |

## 强制工作流

**每一次代码变更都必须通过 myspec 流程在 worktree 中完成，没有例外。**

### myspec agent 完整生命周期

**每个阶段必须通过 Skill 工具调用对应技能，不得跳过或自行替代。**

```
Phase 1: 设计
  Skill: myspec-br（苏格拉底式对话）
  │
  ├── 需要产品决策？→ 派遣专属 pm 子智能体评估 → 接收结论
  ├── 需要技术决策？→ 派遣专属 architect 子智能体评估 → 接收结论
  ├── 综合结论继续苏格拉底式对话
  └── 设计文档完成，进入下一阶段

Phase 2: 方案
  Skill: myspec-propose
  自动生成 proposal / specs / design / tasks 工件

Phase 3: 实施
  Skill: myspec-apply
  在 worktree 中按 task group 编码，每组完成后提交

Phase 4: 审查
  派遣专属 reviewer 子智能体审查代码
  ├── 接收审查报告
  └── 修复问题，重新派遣 reviewer 直到通过

Phase 5: 验证
  Skill: myspec-verify
  ├── 派遣专属 qa 子智能体验证功能
  ├── 接收验证报告
  └── 回填工件使其匹配最终实现

Phase 6: 合并
  myspec agent 通知主会话"就绪待合入" → 等待调度
  主会话通过 SendMessage 通知"可以合入"
  Skill: myspec-merge（含 myspec-catchup + 合入 main）

Phase 7: 归档
  Skill: myspec-merge 内部执行归档
  ├── 同步 delta specs 到 main specs
  ├── 归档 change 目录到 archive/
  └── 清理 worktree 和分支
```

### 合并协调规则

- **串行合入**：同一时间只能有一个 change 合入 main
- **myspec agent 主动请求**：完成验证后，mysagent agent 返回主会话，报告"就绪待合入"
- **主会话调度**：主会话根据当前 main 状态和其他 change 的进度，通过 SendMessage 通知 myspec agent 执行合并
- **合入前必须 catchup**：myspec-merge 技能内部处理同步 main 最新代码
- **合入后立即归档**：myspec-merge 技能内部处理归档
- **冲突由 myspec agent 自行解决**：catchup 发现冲突时，myspec agent 在 worktree 中解决

### 跳过条件（仅此两种情况可跳过 myspec 流程）

| 情况 | 可跳过 | 仍必须执行 |
|------|--------|-----------|
| 修复编译错误 / 拼写错误 / 配置微调 | 整个 myspec 流程 | 直接在 main 上修复并提交 |
| 纯文档修改（只改 docs/ 目录） | myspec 流程 | 直接在 main 上更新 |

### 绝对禁止

- ❌ 跳过 myspec 流程直接在 main 上写业务代码
- ❌ myspec agent 跳过或自行替代 myspec 技能（必须通过 Skill 工具调用）
- ❌ 跳过 reviewer / qa 直接合入
- ❌ 多个 change 同时合入 main
- ❌ 不经 catchup 直接合入
- ❌ 主会话进入设计决策或代码审查
- ❌ 不同 change 的答复团队共享实例（上下文污染）
- ❌ 答复团队只回答被问到的问题，不主动发现缺失
- ❌ reviewer 只检查代码质量，不检查用户路径
- ❌ qa 只验证技术功能，不走真实用户流程

## Agent 派遣规范

### 主会话派遣一组 change 团队

当用户发起一个新 change 时，主会话需要派遣：
1. **myspec agent**（1 个）— 负责完整 myspec 生命周期
2. **答复团队**（多个）— 专属该 myspec agent，由主会话预派遣

myspec agent prompt 必须包含：
1. **问题描述** — 用户提出的问题或需求
2. **全局背景** — 整个系统在做什么，当前功能的完整上下文
3. **change 名称** — kebab-case 标识符
4. **worktree 路径** — 由 myspec-gwt 创建
5. **技能链指引** — 明确要求通过 Skill 工具依次调用 myspec 技能
6. **答复团队名称** — 告知 myspec agent 其专属答复团队的 agent 名称，便于直接交互

答复团队 agent prompt 必须包含：
1. **角色定义** — 你是该 change 的专属 pm/architect/reviewer/qa
2. **全局上下文** — 当前 change 在做什么，系统全貌
3. **交互对象** — 对应 myspec agent 的名称，直接发送结论
4. **主动发现要求** — 明确要求输出"发现的其他问题"

### 并行限制

- 主会话同时管理的 change 团队不超过 3 组（每组 = 1 myspec agent + 答复团队）
- 每个 myspec agent 同时派遣的答复团队子智能体不超过 3 个

## 提交规范

- commit message 使用中文，格式：`type(scope): 描述`
- 类型：feat / fix / refactor / docs / test / chore
- myspec agent 在 worktree 中按 task group 分组提交

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
