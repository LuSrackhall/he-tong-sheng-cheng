# 代码质量提升 + UI/UX 全面优化 设计文档

## Context

资产租赁与催缴管理系统，Go 后端（Gin + GORM）+ Vue 3 前端（自定义 CSS）。支持 SQLite 和 PostgreSQL 双数据库。

当前存在以下问题：
- sqlite/ 和 postgres/ 仓库实现代码 100% 重复（~400 行），因为 GORM 已抽象 SQL 方言，实际无数据库特定逻辑
- AuthHandler 内部重复创建 AuthMiddleware 实例
- 多处静默忽略错误（bcrypt hash、Count 查询等）
- 无优雅关停机制，`os.Exit(0)` 被用作重启
- 混合中英文错误消息（~40 条英文错误）
- config.go 中存在死代码（AIProviderURL/AIProviderKey）
- 前端缺少 Dashboard 概览页、部分 UX 优化空间

## Goals / Non-Goals

**Goals:**
1. 消除 sqlite/ 和 postgres/ 之间的代码重复
2. 修复 AuthMiddleware 重复创建
3. 正确处理所有被忽略的错误
4. 实现优雅关停（graceful shutdown）
5. 统一所有错误消息为中文
6. 删除死代码
7. 新增 Dashboard 概览页 + 后端 API
8. 优化侧边栏分组
9. 优化催缴页面空状态引导

**Non-Goals:**
- 不引入新的前端 UI 框架（保持自定义 CSS）
- 不重构数据库迁移机制
- 不添加新的业务功能
- 不修改认证流程或权限模型

## Decisions

### D1: 仓库代码去重策略

**选择：提取公共代码到 `internal/repository/common/`**

理由：
- GORM 的 API 已经抽象了 SQL 方言差异，当前 sqlite/ 和 postgres/ 的 repo 方法实现 100% 相同
- 只有 `Setup()` 函数需要数据库特定逻辑（驱动导入、DSN 构造）
- 方案简洁，不需要构建标签等额外机制

实现：
- 新建 `internal/repository/common/repos.go`
  - 将所有 repo struct 定义、构造函数、方法实现移入
- `sqlite/setup.go` 和 `postgres/setup.go` 仅保留 `Setup()` 函数
  - 导入 `common` 包，使用其中的构造函数
- 删除 `sqlite/repos.go`, `sqlite/contract.go`, `sqlite/payment.go`, `sqlite/tenant.go`
- 删除 `postgres/repos.go`, `postgres/contract.go`

考虑过的替代方案：
- 构建标签（go:build）：过度复杂化，收益不明显
- 合并为单一实现 + 驱动注入：改变过多，风险高

### D2: AuthMiddleware 复用

**选择：AuthHandler 接收 middleware 实例而非 jwtSecret**

实现：
- `NewAuthHandler(userRepo, authmw)` 替代 `NewAuthHandler(userRepo, jwtSecret)`
- `main.go` 中共享同一个 `authmw` 实例

### D3: 优雅关停

**选择：`http.Server` + `signal.Notify` + `srv.Shutdown(ctx)`**

实现：
- 替换 `r.Run()` 为 `http.Server{Handler: r}` + `srv.ListenAndServe()`
- 监听 SIGINT/SIGTERM
- 收到信号后调用 `srv.Shutdown(ctx)`，timeout 10 秒
- backup.go 的 `os.Exit(0)` 改为通过 channel 通知主函数执行优雅关停

Timeout 选择 10 秒的理由：
- 足够完成 in-flight 请求
- 不会让 Docker/systemd 的 stop timeout（通常 30s）超时
- 是 Go 社区的通用标准

### D4: 错误处理

**选择：逐一修复，将 `_` 替换为显式错误处理**

关键修复点：
- `setup.go:43` — `hash, _ := bcrypt.GenerateFromPassword(...)` → 检查错误，log.Fatal
- `repos.go:41` — `r.db.Model(&domain.Receipt{}).Count(&total)` → 检查并返回错误
- `contract.go` 中 `contractH.DownloadContract` 等处的静默错误

### D5: 死代码清理

- 删除 `config.go` 中的 `AIProviderURL` 和 `AIProviderKey` 字段
- 删除对应的环境变量读取

### D6: 错误消息统一

将所有英文错误消息翻译为中文用户友好消息。主要涉及：
- `auth.go` — ~12 条
- `tenant.go` — ~7 条
- `contract.go` — ~13 条
- `receiptbook.go` — ~3 条
- `main.go` — ~1 条

### D7: Dashboard 概览

**新增后端 API：`GET /api/dashboard/stats`**

响应格式：
```json
{
  "activeContracts": 12,
  "monthlyRevenue": 45000,
  "overdueContracts": 3,
  "newContractsThisMonth": 2
}
```

实现：
- 新建 `internal/transport/handler/dashboard.go`
- 需要的查询：活跃合同数、本月收款总额、逾期合同数、本月新增合同数
- 路由注册在 protected 组下

**前端 Home.vue：**
- 默认路由从 `/new-contract` 改为 `/`
- 展示四个统计卡片
- 使用自定义 CSS，保持与现有风格一致

### D8: 侧边栏优化

**当前状态：** 已有 3 个分组（日常操作、数据管理、系统设置），结构基本合理。

**调整方案：**
- 新增"概览"入口（首页 Dashboard）
- 将"合同管理"从"数据管理"移至"日常操作"（与催缴清单并列）
- 重命名分组：
  - 日常操作 → 业务管理
  - 数据管理 → 基础数据
  - 系统设置 → 系统管理

### D9: 催缴页面空状态

**当前状态：** ArrearsList.vue 已有每个 tab 的空状态提示（"暂无XXX的合同"）。

**优化：** 当所有 tab 计数均为 0 时，显示更醒目的引导："暂无需要催缴的合同。所有合同收款状态正常。" 空状态页面无需添加"请先创建合同"的 CTA，因为此时系统运行正常。

## Risks / Trade-offs

1. **[风险] 仓库重构可能引入编译错误** → 缓解：重构后立即运行 `go test ./...` 和 `go build ./...`
2. **[风险] 优雅关停的 10 秒超时可能不够** → 缓解：PDF 生成等长耗时操作应该在客户端重试，10 秒对 HTTP API 足够
3. **[风险] Dashboard API 查询可能影响性能** → 缓解：使用简单的 COUNT/SUM 查询，有索引支撑，数据量不大
4. **[取舍] 错误消息中文化可能影响日志分析** → 缓解：日志仍保留英文结构化日志，仅 API 返回中文消息
