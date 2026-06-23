# 项目宪法 — 资产租赁与催缴管理系统

## 语言规范

- 所有 git commit message **必须使用中文**
- 与用户交流使用中文
- 代码注释使用中文（仅在必要时添加）

## 团队成员

当任务需要多角色协作时，按以下配置派遣 agent。**所有 agent 默认只读，不直接修改源代码。**

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

## Agent 派遣规则

### 核心原则

1. **审查用 agent，实施由主控执行** — agent 输出方案和审查意见，代码改动由 Claude 直接完成
2. **并行最多 3 个 agent** — tmux 窗格有限，避免互相干扰
3. **每轮改动后立即提交** — 锁定成果，防止 agent 回退
4. **严格限制写入范围** — 每个 agent 有明确的文件边界

### 派遣模式

**小改动（bug fix、配置微调）：**
- 不派遣 agent，直接修改并提交

**中等功能（新页面、新 API）：**
```
pm 评估优先级 → 我执行改动 → reviewer 审查 → 我修复 → 提交
```

**大型功能（新模块、架构变更）：**
```
myspec-br（pm + architect 参与设计）
  → myspec-gwt（创建 worktree 隔离）
    → myspec-apply（我在 worktree 中实施）
      → myspec-verify（qa 验证 + 用户确认）
        → myspec-merge（合并回 main）
```

### 反模式（禁止）

- ❌ agent 直接修改 `internal/domain/`、`internal/transport/`、`cmd/` 下的文件
- ❌ 同时启动超过 3 个 agent
- ❌ 让 agent 同时写入同一个文件
- ❌ 跳过验证直接提交到 main

## 文件保护

以下文件为**高风险区域**，修改前必须先 `git stash`：
- `internal/domain/` — 领域模型和仓库接口
- `internal/di/` — 依赖注入
- `cmd/server/main.go` — 路由注册
- `internal/config/` — 配置管理

## 开发工作流

### 快速迭代模式（默认）

```
1. 用户提出需求
2. 我评估复杂度，决定是否需要 agent
3. 如需要 → 派遣 pm/architect 出方案
4. 我执行代码改动
5. 运行 go test + vue-tsc 验证
6. git commit（中文 message）
7. 如需要 → 派遣 reviewer 审查
8. 修复审查发现的问题
9. 交付
```

### 发布前检查清单

- [ ] `go test ./... -count=1` 全部通过
- [ ] `go build ./...` 编译通过
- [ ] `vue-tsc --noEmit` 类型检查通过
- [ ] `npm run build` 前端构建通过
- [ ] 启动服务验证健康检查
- [ ] 端到端登录测试

## 项目结构速查

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
docs/                        # 用户手册 + 部署指南 + API 文档
tests/                       # E2E 测试 + 集成测试
```
