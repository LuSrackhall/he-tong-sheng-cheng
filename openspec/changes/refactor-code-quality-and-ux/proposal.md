## Why

资产租赁管理系统存在代码质量和用户体验方面的技术债务：sqlite/ 和 postgres/ 仓库实现之间约 400 行完全重复的代码，AuthMiddleware 被重复创建，多处错误被静默忽略，无优雅关停机制，错误消息中英文混杂。这些问题影响可维护性、可靠性和用户体验。需要系统性修复以达到生产级水准。

## What Changes

1. **仓库代码去重**：将 sqlite/ 和 postgres/ 中 100% 相同的 repo 实现提取到 `internal/repository/common/`，两个驱动包仅保留各自的 `Setup()` 函数
2. **AuthMiddleware 复用**：AuthHandler 不再自行创建 AuthMiddleware，改为接收主程序共享的实例
3. **错误处理改进**：修复所有被 `_` 忽略的错误（bcrypt hash、Count 查询等），改为正确处理
4. **优雅关停**：替换 `r.Run()` 为 `http.Server` + SIGINT/SIGTERM 信号监听 + 10 秒超时的 `srv.Shutdown(ctx)`
5. **os.Exit(0) 替换**：backup.go 中恢复数据库后的 `os.Exit(0)` 改为通知优雅关停流程
6. **错误消息统一**：将所有 ~40 条英文错误消息翻译为中文用户友好消息
7. **死代码清理**：删除 config.go 中未使用的 AIProviderURL 和 AIProviderKey
8. **Dashboard 概览**：新增后端 `GET /api/dashboard/stats` 端点和前端 Home.vue 概览页，展示活跃合同数、本月收款、逾期合同数、本月新增合同数
9. **侧边栏优化**：重命名分组标签（业务管理、基础数据、系统管理），新增首页入口
10. **催缴空状态优化**：当所有催缴 tab 均为空时显示更醒目的引导信息

## Capabilities

### New Capabilities
- `dashboard-stats`: 后端 Dashboard 统计 API 端点和前端概览页面
- `graceful-shutdown`: 服务器优雅关停机制，替代直接 os.Exit

### Modified Capabilities
- `template-driven-contract-form`: 错误消息统一为中文

## Impact

**受影响的代码：**
- `internal/repository/sqlite/` — 删除大部分文件，仅保留 setup.go
- `internal/repository/postgres/` — 删除大部分文件，仅保留 setup.go
- `internal/repository/common/` — 新建，接收所有 repo 实现
- `internal/transport/handler/` — auth.go、tenant.go、contract.go、receiptbook.go、backup.go、新增 dashboard.go
- `internal/transport/middleware/` — 无变化
- `internal/config/config.go` — 删除死代码
- `cmd/server/main.go` — 优雅关停、路由注册
- `frontend/src/App.vue` — 侧边栏分组调整
- `frontend/src/views/Home.vue` — 新建 Dashboard 概览页
- `frontend/src/views/ArrearsList.vue` — 空状态优化
- `frontend/src/router/index.ts` — 默认路由改为 `/`

**API 变化：**
- 新增 `GET /api/dashboard/stats`（需认证）

**风险：**
- 仓库重构需谨慎，确保 `go test ./...` 全部通过
- 优雅关停 timeout 10 秒对 PDF 生成等长操作可能不够，但对 HTTP API 足够
