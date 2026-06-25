# GitHub Actions CI/CD 设计文档

## Context

### 当前状态

- **技术栈**：Go 1.26.1 + Vue 3 + TypeScript + SQLite/PostgreSQL
- **构建工具**：Makefile（`build`、`test` 命令）、Dockerfile（多阶段构建）
- **代码检查**：`go test`、`vue-tsc`、`npm run build`（无 Lint）
- **现状**：无 CI/CD，所有检查依赖开发者本地手动执行

### 问题

- 开发者可能忘记跑测试就提交
- PR 审查无法自动验证代码质量
- 多人协作时，main 分支可能被破坏性提交污染

---

## Goals / Non-Goals

### Goals

- 在 GitHub 上实现自动 CI 检查
- 每次 push/PR 自动运行：Go 测试、前端类型检查、构建验证
- 支持 SQLite + PostgreSQL 双数据库测试
- 仅报告，不阻断合并（PR 显示红叉但允许手动合并）

### Non-Goals

- 不涉及自动部署（CD）
- 暂不添加 ESLint（未安装）
- 暂不添加安全扫描

---

## Decisions

### 1. CI 方案选择：多 Job 并行

**选择**：方案 2 - 多 Job 并行

**理由**：
- 前端和后端检查完全独立，互不阻塞
- SQLite 和 PostgreSQL 测试并行，节省时间
- 配置清晰，易于维护
- 某项失败不影响其他 job

**替代方案**：
- 方案 1（单一 Job）：所有检查串行，总耗时长
- 方案 3（矩阵策略）：配置简洁，但会重复执行前端检查

### 2. PostgreSQL 测试切换机制

**选择**：通过 `MODE` 环境变量切换

**依据**：
- `internal/config/config.go` 中 `MODE` 字段控制数据库选择
- `MODE=postgres` 使用 PostgreSQL，默认使用 SQLite

**CI 中 PostgreSQL 测试需要的环境变量**：
```
MODE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=design_platform
JWT_SECRET=test-secret
```

### 3. 前端检查策略

**选择**：只运行 `npm run build`（已含 vue-tsc）

**理由**：
- `package.json` 中 `"build": "vue-tsc -b && vite build"`
- 无需单独运行 `vue-tsc --noEmit`，避免重复

### 4. Lint 策略

**选择**：暂不添加

**理由**：
- ESLint 未安装，需要配置规则文件
- 先把基础检查跑起来，后续需要时再加

### 5. Go 版本管理

**选择**：使用 `go-version-file: 'go.mod'`

**理由**：
- 自动从 go.mod 读取版本，无需硬编码
- go.mod 更新时 CI 自动跟随

### 6. CGO 配置

**选择**：显式设置 `CGO_ENABLED: "1"`

**理由**：
- 项目使用 `github.com/mattn/go-sqlite3`（CGO 依赖）
- Makefile 中明确设置了 `CGO_ENABLED=1`
- Ubuntu runner 默认有 gcc，但 CGO 开关需显式声明

### 7. JWT_SECRET 配置

**选择**：两个 backend job 都设置 `JWT_SECRET: test-secret`

**理由**：
- `config.go` 中 `JWT_SECRET` 是必填项，缺失时程序 `log.Fatal` 退出
- 测试可能触发配置加载，需要提供测试值

---

## Risks / Trade-offs

### 风险 1：PostgreSQL 测试实际增量价值有限

**现状**：当前测试文件主要是纯计算函数测试，无 repository 层单元测试

**影响**：backend-postgres job 运行的 `go test ./...` 实际执行的测试与 backend-sqlite 完全相同

**缓解**：作为基础设施预留是合理的，后续添加 repository 测试时自然会产生差异

### 风险 2：ESLint 未配置

**现状**：`frontend/package.json` 中无 ESLint 相关依赖

**影响**：无法进行前端代码风格检查

**缓解**：先跑基础检查，后续需要时再配置 ESLint

### 风险 3：`.env.example` 与代码不一致

**现状**：`DB_NAME` 在 `.env.example` 中为 `asset_leasing`，在 `config.go` 中默认为 `design_platform`

**影响**：不影响 CI（CI 直接使用代码默认值），但会误导开发者

**缓解**：后续统一

---

## 最终 CI Workflow 配置

```yaml
name: CI

on:
  push:
    branches: ['*']
  pull_request:
    branches: ['*']

jobs:
  frontend-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      - run: cd frontend && npm ci && npm run build

  backend-sqlite:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - run: go test ./... -count=1
        env:
          JWT_SECRET: test-secret
          CGO_ENABLED: "1"
      - run: go build ./...
        env:
          CGO_ENABLED: "1"

  backend-postgres:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: design_platform
        ports:
          - 5432:5432
        options: --health-cmd pg_isready --health-interval 10s --health-timeout 5s --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - run: go test ./... -count=1
        env:
          MODE: postgres
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASS: postgres
          DB_NAME: design_platform
          JWT_SECRET: test-secret
          CGO_ENABLED: "1"
      - run: go build ./...
        env:
          CGO_ENABLED: "1"
```

---

## 答复团队评估汇总

### PM 评估：✅ 通过

**产品价值**：
- 质量保障：每次 push 自动验证前后端构建+测试
- 双数据库兼容：同时验证 SQLite（开发）和 PostgreSQL（生产）
- 零成本启动：从 0 到 1 的质变
- 快速反馈：并行 job 设计，几分钟内出结果

**建议**：
- Go 版本使用 `go-version-file: 'go.mod'` 代替硬编码

### Architect 评估：✅ 通过

**技术可行性**：
- 所有环境变量与 `config.go` 完全匹配
- 前端 build 包含类型检查，无需单独加 `vue-tsc --noEmit`

**修正建议**：
- backend-sqlite 需要 `JWT_SECRET: test-secret` 环境变量
- 需要显式设置 `CGO_ENABLED: "1"`

### Reviewer 评估：✅ 通过

**一致性**：与 CLAUDE.md 发布前检查清单完全一致

**完整性**：覆盖所有可自动化的检查项

**修正建议**：
- CGO_ENABLED=1 必须显式设置
- Go 版本应与 go.mod 对齐（1.26.1）
