## 1. 认证安全

- [x] 1.1 在 `internal/config/config.go` 中新增 `ADMIN_PASSWORD` 和 `DBSSLMode` 字段，ADMIN_PASSWORD 未设置时 `log.Fatalf`
- [x] 1.2 修改 `internal/transport/middleware/auth.go` 的 `ParseToken`，在 keyFunc 中校验签名方法为 HMAC 系列
- [x] 1.3 新建 `internal/security/ratelimit.go`，实现 `LoginRateLimiter`（内存 map + mutex，5 分钟/5 次限制，后台清理 goroutine）
- [x] 1.4 修改 `internal/transport/handler/auth.go`，注入 rate limiter，登录失败时记录、成功时重置，超限时返回 429
- [x] 1.5 修改 `cmd/server/main.go`，创建 rate limiter 实例并注入 AuthHandler，启动 cleanup goroutine

## 2. 数据安全

- [x] 2.1 在 `internal/transport/handler/tenant.go` 中定义 `maskIDCard()` 辅助函数，List 接口返回时对 IDCard 脱敏
- [x] 2.2 在 `cmd/server/main.go` 中设置 `r.MaxMultipartMemory = 10 << 20`
- [x] 2.3 修改 `internal/repository/postgres/setup.go`，接收 sslmode 参数，拼接到 DSN 中
- [x] 2.4 修改 `internal/di/deps.go`，传递 sslmode 到 postgres.Setup()
- [x] 2.5 修改 `internal/transport/handler/backup.go` 的 Restore 方法，从 POST body 读取 confirmed 字段
- [x] 2.6 修改 `internal/transport/handler/payment.go` 的 VoidPayment，将 `err.Error()` 替换为通用消息，内部错误 log 记录

## 3. 基础设施安全

- [ ] 3.1 修改 `internal/repository/sqlite/setup.go`，在 gorm.Open 后执行 WAL/foreign_keys/busy_timeout PRAGMA
- [ ] 3.2 修改 `internal/repository/sqlite/setup.go` 和 `internal/repository/postgres/setup.go`，接收 password 参数，检查 bcrypt 错误
- [ ] 3.3 删除 `internal/repository/sqlite/repos.go:202` 的 `var _ *gorm.DB = nil`

## 4. 前端适配

- [ ] 4.1 修改 `frontend/src/api/index.ts`，备份恢复接口 confirmed 从 URL 移到 FormData
- [ ] 4.2 修改 `frontend/src/views/Login.vue`，处理 429 状态码显示限流提示

---

## Post-Implementation Workflow

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run `/opsx:verify` to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, go to main branch and merge (must ask user)
4. **Archive**: Run `/opsx:archive` on main
5. **Cleanup**: `git worktree remove .worktrees/change/<name>`
