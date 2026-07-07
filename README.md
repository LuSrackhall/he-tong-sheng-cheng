# 资产租赁与催缴管理系统

村/社区资产（商铺、车位、摊位、设备）租赁合同管理、租金收缴与欠费催缴一体化系统。

## 技术栈

- **后端:** Go + Gin + GORM
- **前端:** Vue 3 (Composition API) + TypeScript + Pinia + Vue Router
- **数据库:** SQLite（本地单文件）/ PostgreSQL（服务器）
- **TCA:** Knowledge Runtime + Capability DSL + Constitution Guard
- **E2E:** Playwright 自动化测试
- **Agent 接口:** MCP Server
- **构建:** Vite -> `go:embed` 单一二进制部署

## 快速启动

### 本地开发（SQLite 模式）

```bash
# 1. 构建前端（输出到 cmd/server/dist/）
cd frontend && npm install && npm run build && cd ..

# 2. 编译并启动服务
JWT_SECRET=dev-secret ADMIN_PASSWORD=admin123 go run ./cmd/server

# 服务运行在 http://localhost:8080
# 默认管理员账号: admin / admin123
```

### Docker 部署（PostgreSQL 模式）

```bash
# 1. 设置环境变量
export JWT_SECRET="your-secure-jwt-secret"
export ADMIN_PASSWORD="your-secure-admin-password"

# 2. 启动服务
docker compose up -d

# 3. 查看日志
docker compose logs -f app
```

### Make 命令

```bash
make build          # 构建前端+后端
make dev            # 开发模式运行
make test           # 运行所有测试
make test-cover     # 运行测试并生成覆盖率报告
make lint           # 代码质量检查（go vet + vue-tsc）
make typecheck      # 类型检查（Go + TypeScript）
make docker-build   # 构建 Docker 镜像
make docker-up      # 启动 Docker Compose 服务
make docker-down    # 停止 Docker Compose 服务
make clean          # 清理构建产物
```

## 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `JWT_SECRET` | 是 | - | JWT 签名密钥 |
| `ADMIN_PASSWORD` | 是 | - | 初始管理员密码 |
| `MODE` | 否 | `sqlite` | 数据库模式：`sqlite` 或 `postgres` |
| `DB_HOST` | 否 | `localhost` | PostgreSQL 主机地址 |
| `DB_PORT` | 否 | `5432` | PostgreSQL 端口 |
| `DB_NAME` | 否 | `asset_leasing` | PostgreSQL 数据库名 |
| `DB_PASS` | 否 | - | PostgreSQL 密码 |
| `PORT` | 否 | `8080` | 服务端口 |

也支持命令行参数（`--mode`, `--db-host` 等），环境变量优先。

## 功能概览

### 三大入口
- **签新合同** -- 分步表单：选资产 -> 录租户 -> 定合同 -> 预览
- **收租金** -- 合同搜索 + 收款弹窗 + 还差多少
- **催缴清单** -- 五级分级（应缴预警/近期提醒/逾期催收/到期预警/欠费追缴）

### 后台管理
- 资产管理、租户管理、合同管理
- 收据本管理、用户管理（管理员/操作员角色分权）
- 系统设置（合同模板字段映射）

## 项目结构

```
cmd/server/main.go          # 入口 + 路由 + go:embed
internal/
  config/                    # CLI 参数解析 + 环境变量
  di/                        # 依赖注入
  domain/                    # 实体 + Repository 接口
    calc/                    # 纯函数计算引擎（已测）
  repository/
    sqlite/                   # SQLite 实现
    postgres/                 # PostgreSQL 实现
  transport/
    handler/                 # HTTP handlers
    middleware/               # JWT 认证 + SPA fallback
frontend/
  src/
    views/                   # 12 个 Vue 页面
    api/                     # Axios 封装 + 接口类型
    stores/                  # Pinia 状态管理
    router/                  # 路由配置
tests/
  api_test.go               # API 集成测试（Go）
  calc_test.go              # 计算引擎测试（Go）
  e2e.sh                    # 端到端 bash 测试脚本
```

## 开发

```bash
# 后端
make test                    # 运行测试
make test-cover              # 测试覆盖率

# 前端
cd frontend
npm run dev                  # Vite 开发服务器
npx vue-tsc --noEmit        # TypeScript 类型检查
npm run build                # 生产构建
```

## 测试

```bash
# Go 单元测试 + 集成测试
make test

# API 集成测试（需要先启动服务）
JWT_SECRET=test ADMIN_PASSWORD=admin123 go run ./cmd/server &
bash tests/e2e.sh
```
---
## TCA（Task-Centric Architecture）

系统从 Phase 1 引入了 TCA——以业务能力为中心的质量与运行时体系。核心目录:
- `runtime/` — Knowledge Runtime（加载/解析/规划/验证）
- `knowledge/` — 业务能力 DSL（Capabilities / Rules / Workflows）
- `system/constitution.md` — 13 条不可违反系统公理
- `adapters/` — 运行时适配器（MCP Server / Exploration）

### kr CLI 快速入门

```bash
# 编译 CLI
go build -o kr ./runtime/cmd/kr/

# 预览业务能力执行计划
./kr plan collect-rent

# 预览复合流程
./kr plan sign-new-contract

# 执行能力（生成 Trace）
./kr run collect-rent --validate=false

# 回溯查看执行轨迹
./kr explain <trace-id>
```

**已定义的业务能力：** login / collect-rent / create-contract / issue-receipt / backup-database / create-user / ensure-contract-active

### MCP Server（Agent 接口）

```bash
# 启动（后端需在 :8080 运行）
go run ./adapters/mcp/server.go

# 查看可用工具（6个）
curl -X POST http://localhost:9090/mcp \
  -H "Content-Type: application/json" \
  -d '{"method":"tools/list","id":1}'

# 调用 Dashboard 统计
curl -X POST http://localhost:9090/mcp \
  -H "Content-Type: application/json" \
  -d '{"method":"tools/call","id":2,"params":{"name":"get_dashboard","arguments":{}}}'
```

### Playwright E2E

```bash
# 先启动服务
JWT_SECRET=dev-secret ADMIN_PASSWORD=admin123 go run ./cmd/server &

# 运行 E2E 测试（26 个场景）
cd frontend && npx playwright test --config=e2e/playwright.config.ts
```

## 项目结构

```
cmd/server/main.go          # 入口 + 路由
internal/                    # 传统后端分层
  config/ di/ domain/ calc/
  repository/ transport/
runtime/                     # TCA Knowledge Runtime
  cmd/kr/                    # CLI 入口
  internal/                  # model/loader/resolver/planner/guard/trace
knowledge/                   # 业务知识层
  capabilities/workflows/rules/
adapters/                    # TCA Runtime Adapters
  mcp/ exploration/
frontend/                    # Vue 3
  src/views/ e2e/            # 12 页面 + Playwright 测试
system/                      # 系统宪法
tests/                       # 集成测试
```

## 测试速查

```bash
# Go 全量测试
go test ./... -count=1

# TCA 测试
go test ./runtime/... -count=1
bash tests/runtime_integration.sh

# E2E 测试（需先启动服务）
cd frontend && npx playwright test --config=e2e/playwright.config.ts

# 静态分析
go vet ./...
```
