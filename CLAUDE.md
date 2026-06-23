# 项目宪法 — 资产租赁与催缴管理系统

## 语言规范

- 所有 git commit message **必须使用中文**
- 与用户交流使用中文
- 代码注释使用中文（仅在必要时添加）

## 团队成员

| Agent 名称 | 角色 | 职责 | 权限 |
|-----------|------|------|------|
| `pm` | 产品经理 | 需求分析、优先级排序、功能完整性评估、验收标准 | 只读全部 |
| `architect` | 架构师 | 技术方案设计、API 设计、代码审查、架构决策 | 只读全部 |
| `qa` | QA 工程师 | 测试方案、功能验证、回归测试、边界用例 | 只读 + 写 `tests/` |
| `doc-writer` | 文档工程师 | 用户手册、API 文档、部署指南、变更日志 | 只写 `docs/` |
| `reviewer` | 综合审查员 | 多维度代码审查（安全/性能/一致性/可维护性） | 只读全部 |

### 临时成员（按需启用）

| Agent 名称 | 触发场景 |
|-----------|---------|
| `domain-expert` | 新业务领域功能（如催缴策略调整、计算规则变更） |
| `perf-analyst` | 性能瓶颈排查、数据库查询优化 |
| `refactor-specialist` | 大规模重构（如 service 层抽取、代码去重） |

## 强制工作流

**每一次代码变更都必须经过以下流程，没有例外。**

### 流程步骤

```
步骤 1: 派遣 pm agent → 评估需求优先级，输出验收标准
步骤 2: 派遣 architect agent → 输出技术方案（涉及哪些文件、API 设计、数据模型变更）
步骤 3: 我执行代码改动
步骤 4: 派遣 reviewer agent → 审查改动，输出发现的问题
步骤 5: 修复 reviewer 发现的问题
步骤 6: 派遣 qa agent → 验证功能正确性
步骤 7: 派遣 doc-writer agent → 更新受影响的文档
步骤 8: 提交
```

### 跳过条件（仅此三种情况可跳过步骤 1-2）

| 情况 | 可跳过 | 仍必须执行 |
|------|--------|-----------|
| 用户已明确给出完整技术方案（指定了文件、函数、修改内容） | 步骤 1-2 | 步骤 3-8 |
| 修复编译错误（代码已正确，只是拼写/导入问题） | 步骤 1-4 | 步骤 3, 8 |
| 纯文档修改（只改 docs/ 目录） | 步骤 1-5 | 步骤 3, 7, 8 |

### 绝对禁止

- ❌ 跳过 reviewer 直接提交代码改动
- ❌ 跳过 qa 直接提交功能变更
- ❌ 认定改动"太小不需要审查"——没有这种例外
- ❌ agent 直接修改 `internal/` 或 `cmd/` 或 `frontend/src/` 下的源代码文件（agent 只读，只输出方案和审查意见）
- ❌ 同时启动超过 3 个 agent
- ❌ 让多个 agent 同时写入同一个文件

## Agent 派遣规范

### 并行限制
- 同时最多 3 个 agent
- tmux 窗格不足时串行派遣

### 写入权限
- **所有 agent 默认只读**，不直接修改源代码
- agent 输出方案/审查意见，代码改动由我执行
- 唯一例外：qa 可写 `tests/`，doc-writer 可写 `docs/`

### 文件保护
以下文件修改前必须先 `git stash`：
- `internal/domain/` — 领域模型和仓库接口
- `internal/di/` — 依赖注入
- `cmd/server/main.go` — 路由注册
- `internal/config/` — 配置管理

## 提交规范

- 每轮改动后立即提交，锁定成果
- commit message 使用中文，格式：`type(scope): 描述`
- 类型：feat / fix / refactor / docs / test / chore

## 开发工作流（myspec）

大型功能使用 myspec 流程：
```
myspec-br（需求分析+设计文档）
  → myspec-gwt（创建 worktree 隔离）
    → myspec-apply（在 worktree 中实施）
      → myspec-verify（验证+用户确认）
        → myspec-merge（合并回 main）
```

## 发布前检查清单

- [ ] `go test ./... -count=1` 全部通过
- [ ] `go build ./...` 编译通过
- [ ] `vue-tsc --noEmit` 类型检查通过
- [ ] `npm run build` 前端构建通过
- [ ] 启动服务验证健康检查
- [ ] 端到端登录测试

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
frontend/src/
  views/                     # Vue 页面
  api/                       # Axios 封装
  stores/                    # Pinia 状态管理
  router/                    # 路由配置
  composables/               # Vue 组合式函数
docs/                        # 用户手册 + 部署指南 + API 文档
tests/                       # E2E 测试 + 集成测试
```
