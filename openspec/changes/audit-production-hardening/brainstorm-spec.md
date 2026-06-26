## Context

资产租赁与催缴管理系统已完成安全加固（密码强制配置、JWT 算法校验、登录限流、PII 脱敏、请求限制、CSRF 防护、SQLite PRAGMA、优雅关停等）。在正式生产部署前，需全面审计并修复所有影响生产落地的问题。

审计覆盖 Go 后端（42 个文件）、Vue 3 前端（19 个文件）、Dockerfile、docker-compose.yml、数据库 schema。

## Goals / Non-Goals

**Goals:**
- 修复所有 P0 安全/数据完整性问题（5 个）
- 修复所有 P1 健壮性/错误处理问题（4 个）
- 修复所有 P2 性能/可靠性问题（4 个）
- 所有修复通过 `go build ./...`、`go test ./... -count=1`、`npm run build`
- 不破坏现有功能

**Non-Goals:**
- P3 前端重构（Settings.vue 拆分、全局错误边界组件）
- 新增功能
- 数据库 schema 大规模改动（仅添加索引）

## Decisions

### P0 — 安全性 / 数据完整性

**D1: VACUUM INTO 路径注入** (`internal/transport/handler/backup.go:69`)
- `VACUUM INTO ?` 的 GORM 参数化在 SQLite 中不可靠，可能生成错误 SQL
- 修复：使用 `fmt.Sprintf("VACUUM INTO '%s'", backupPath)` 直接拼接路径（路径由服务端生成，不接受用户输入）
- 备选方案 `h.db.Exec("VACUUM INTO " + backupPath)` 有注入风险，不采用

**D2: 恢复后 DB 连接已关闭** (`internal/transport/handler/backup.go:170-171`)
- `sqlDB.Close()` 后延迟 1 秒内有并发请求会 panic
- 修复：在调用 shutdownFn 前，不再手动关闭 sqlDB，由优雅关停流程统一处理。删除 `sqlDB.Close()` 调用

**D3: 租户 Get 接口泄露完整身份证号** (`internal/transport/handler/tenant.go:78-92`)
- List 接口已脱敏，但 Get 接口返回完整 IDCard
- 修复：Get 接口也对 IDCard 脱敏，与 List 一致。编辑场景不需要前端持有完整身份证号

**D4: BackupInfo 暴露服务器路径** (`internal/transport/handler/backup.go:36`)
- `info["path"] = h.dbPath` 泄露内部文件路径
- 修复：移除 `path` 字段，仅返回 `type`、`size`、`lastModified`

**D5: CORS 缺失**
- 生产环境无 CORS 控制
- 修复：新增 `CORS_ORIGINS` 环境变量配置，使用 `github.com/gin-contrib/cors` 中间件。默认空值表示不启用 CORS（同源模式）

### P1 — 健壮性 / 错误处理

**D6: 类型断言 panic 风险** (`internal/transport/handler/auth.go:221` 及其他)
- `userID.(uint)` 无 ok-check，异常时 panic
- 修复：所有 `c.Get("userID")` 的类型断言改为 ok-check 模式，失败返回 500

**D7: VoidPayment 错误状态码** (`internal/transport/handler/payment.go:165`)
- 系统错误时返回 400 而非 500
- 修复：区分业务错误（400 Bad Request）和系统错误（500 Internal Server Error）

**D8: Recovery 中间件日志泄露** (`cmd/server/main.go:73`)
- `gin.Recovery()` 默认打印完整堆栈到日志
- 修复：使用 `gin.CustomRecovery` 仅记录错误消息，不输出堆栈

**D9: 前端 401 处理不完整** (`frontend/src/api/index.ts:18`)
- `window.location.hash = '#/login'` 不重置前端状态
- 修复：改为 `localStorage.removeItem('token'); window.location.reload()`

### P2 — 性能 / 可靠性

**D10: SQLite 缺少连接池配置** (`internal/repository/sqlite/setup.go`)
- 未配置 `SetMaxOpenConns` 等参数
- 修复：添加 `SetMaxOpenConns(1)` (SQLite 单写)、`SetMaxIdleConns(1)`、`SetConnMaxLifetime(0)`

**D11: ArrearsList 无分页** (`internal/transport/handler/arrears.go:34`)
- `ListUnpaid()` 加载全部未付合同到内存，数据量大时 OOM
- 修复：为 `ListUnpaid` 添加分页支持，前端 ArrearsList 适配分页参数

**D12: 数据库索引缺失**
- `contracts.status`、`payments.contract_id`、`payments.voided`、`receipts.payment_id` 常用查询字段无索引
- 修复：在 domain model 的 GORM tag 中添加 `gorm:"index"`

**D13: Dockerfile 以 root 运行** (`Dockerfile:25`)
- 生产最佳实践应使用非 root 用户
- 修复：添加 `RUN adduser -D appuser` + `USER appuser`，确保目录权限正确

### 配置变更

**D14: 新增 CORSOrigins 配置项** (`internal/config/config.go`)
- 新增字段 `CORSOrigins string`
- 来源：环境变量 `CORS_ORIGINS`，默认空

## Risks / Trade-offs

- **SQLite MaxOpenConns(1)**：并发写入排队，但 SQLite 本身就是单写模型，这是正确配置 → 无实际风险
- **ArrearsList 分页**：前端需适配分页参数 → 改动量小，复用现有 parsePagination
- **CORS 默认不启用**：开发环境需要手动配置 → 生产安全优先，开发可通过环境变量开启
- **Get 接口身份证脱敏**：编辑场景前端看不到完整身份证号 → 但用户本身知道自己的身份证号，脱敏不影响编辑
