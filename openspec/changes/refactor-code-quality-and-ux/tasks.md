## 1. 后端代码质量修复

- [x] 1.1 创建 `internal/repository/common/repos.go`，将 sqlite/ 和 postgres/ 中所有重复的 repo struct 定义、构造函数和方法实现移入
- [x] 1.2 精简 `internal/repository/sqlite/setup.go`，删除所有 repo struct 和方法，仅保留 Setup() 函数，使用 common 包的构造函数
- [x] 1.3 精简 `internal/repository/postgres/setup.go`，同上
- [x] 1.4 删除 sqlite/ 中不再需要的文件：repos.go、contract.go、payment.go、tenant.go
- [x] 1.5 删除 postgres/ 中不再需要的文件：repos.go、contract.go
- [x] 1.6 更新 `internal/di/` 包，改用 common 包的 repo 构造函数
- [x] 1.7 修复 AuthMiddleware 重复创建：AuthHandler 接收 `*middleware.AuthMiddleware` 而非 jwtSecret
- [x] 1.8 修复所有被 `_` 忽略的错误（setup.go 的 bcrypt hash、repos.go 的 Count 查询等）
- [x] 1.9 统一所有 handler 错误消息为中文（auth.go、tenant.go、contract.go、receiptbook.go 等）
- [x] 1.10 删除 config.go 中的 AIProviderURL 和 AIProviderKey 死代码

## 2. 优雅关停

- [x] 2.1 在 main.go 中替换 `r.Run()` 为 `http.Server` + `os/signal` 监听 + `srv.Shutdown(ctx)` 实现优雅关停
- [x] 2.2 修改 BackupHandler 接收 shutdownFn 回调，替换 `os.Exit(0)` 为调用 shutdownFn 触发优雅关停
- [x] 2.3 在 main.go 中传入 shutdownFn 给 BackupHandler

## 3. Dashboard API

- [ ] 3.1 在 domain 层定义 DashboardStats 结构体
- [ ] 3.2 在 ContractRepo 和 PaymentRepo 接口中添加 Dashboard 所需的查询方法
- [ ] 3.3 在 common/repos.go 中实现 Dashboard 查询方法（CountActive、MonthlyRevenue、CountOverdue、CountNewThisMonth）
- [ ] 3.4 创建 `internal/transport/handler/dashboard.go`，实现 DashboardStatsHandler
- [ ] 3.5 在 main.go 中注册 `GET /api/dashboard/stats` 路由

## 4. 前端 UI/UX 优化

- [ ] 4.1 修改默认路由从 `/new-contract` 改为 `/`，创建 Home.vue Dashboard 概览页
- [ ] 4.2 在 App.vue 中新增"概览"菜单入口，重命名侧边栏分组标签
- [ ] 4.3 优化 ArrearsList.vue 空状态：当所有 tab 均为空时显示醒目引导

## 5. 构建验证

- [ ] 5.1 运行 `go build ./...` 确认编译通过
- [ ] 5.2 运行 `go test ./... -count=1` 确认测试通过
- [ ] 5.3 运行 `npm run build`（在 frontend/ 目录）确认前端构建通过

---

## Post-Implementation Workflow

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run myspec-verify to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, notify team-lead for merge scheduling
4. **Archive**: Run myspec-merge after receiving merge approval
5. **Cleanup**: worktree and branch cleanup handled by merge skill
