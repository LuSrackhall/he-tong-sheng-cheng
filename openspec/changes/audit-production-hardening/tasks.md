## 1. 后端安全性修复

- [x] 1.1 修复 VACUUM INTO 路径构造：backup.go:69 改用 `fmt.Sprintf("VACUUM INTO '%s'", backupPath)`
- [x] 1.2 移除 Restore 手动关闭 DB 连接：backup.go 删除 `sqlDB.Close()` 调用
- [x] 1.3 移除 BackupInfo 暴露路径：backup.go 删除 `info["path"] = h.dbPath`
- [x] 1.4 租户 Get 接口身份证号脱敏：tenant.go Get 方法返回前调用 `maskIDCard()`
- [x] 1.5 新增 CORS 配置项：config.go 添加 `CORSOrigins` 字段（来源 CORS_ORIGINS 环境变量）
- [x] 1.6 新增 CORS 中间件：main.go 在路由前按需启用 gin-contrib/cors
- [x] 1.7 安装 gin-contrib/cors 依赖：`go get github.com/gin-contrib/cors`

## 2. 后端健壮性修复

- [x] 2.1 提取类型断言 helper 函数：在 handler 包中新增 `getUintFromContext(c *gin.Context, key string) (uint, error)`
- [x] 2.2 修复 auth.go ChangePassword 的类型断言：使用 helper 函数替换 `userID.(uint)`
- [x] 2.3 修复 auth.go DeleteUser 的类型断言：使用 helper 函数
- [x] 2.4 修复 VoidPayment 状态码：系统错误返回 500 而非 400
- [x] 2.5 自定义 Recovery 中间件：main.go 替换 `gin.Recovery()` 为 `gin.CustomRecovery`

## 3. 后端性能修复

- [x] 3.1 SQLite 连接池配置：sqlite/setup.go 添加 `SetMaxOpenConns(1)` 等参数
- [x] 3.2 ArrearsList 分页 - Repository 层：repos.go 的 ListUnpaidPaging 方法
- [x] 3.3 ArrearsList 分页 - Handler 层：arrears.go 使用 parsePagination
- [x] 3.4 添加数据库索引 tag：contract.go status、payment.go contractId/voided、receipt.go paymentId

## 4. 前端修复

- [x] 4.1 修复 401 处理：api/index.ts 改为 `localStorage.removeItem + window.location.reload()`
- [x] 4.2 ArrearsList 前端分页适配：后端分页对前端透明，无需前端改动

## 5. 部署修复

- [x] 5.1 Dockerfile 非 root 用户：添加 appuser 并设置目录权限

## 6. 验证

- [ ] 6.1 `go build ./...` 编译通过
- [ ] 6.2 `go test ./... -count=1` 测试通过
- [ ] 6.3 `npm run build` 前端构建通过


## Post-Implementation Workflow

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run `/opsx:verify` to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, go to main branch and merge (must ask user)
4. **Archive**: Run `/opsx:archive` on main
5. **Cleanup**: `git worktree remove .worktrees/change/<name>`

**Iteration**: If user does not accept, analyze the issue and recommend:
fix in place / new change / git reset + stash / git reset / abandon.
