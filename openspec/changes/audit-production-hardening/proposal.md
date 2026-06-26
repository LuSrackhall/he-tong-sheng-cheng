## Why

资产租赁与催缴管理系统已完成安全加固，但全面审计发现 13 个影响生产落地的问题，包括路径注入、连接泄漏、PII 泄露、缺失 CORS、类型断言 panic、错误状态码、SQLite 缺少连接池、催缴列表无分页等。需在生产部署前统一修复。

## What Changes

- **修复 VACUUM INTO SQL 路径注入**：backup.go 中改用 `fmt.Sprintf` 构造安全路径
- **修复恢复后 DB 连接泄漏**：删除手动 `sqlDB.Close()`，由优雅关停统一处理
- **租户 Get 接口身份证号脱敏**：与 List 接口保持一致，Get 也返回脱敏后 IDCard
- **移除 BackupInfo 暴露路径**：不再返回 `dbPath` 字段
- **新增 CORS 中间件**：通过 `CORS_ORIGINS` 环境变量配置，使用 gin-contrib/cors
- **修复类型断言 panic**：所有 `c.Get("userID")` 改为 ok-check 模式
- **修正 VoidPayment 错误状态码**：系统错误返回 500 而非 400
- **自定义 Recovery 中间件**：仅记录错误消息，不输出堆栈
- **前端 401 处理改为刷新页面**：`window.location.reload()` 重置状态
- **SQLite 连接池配置**：`SetMaxOpenConns(1)`、`SetMaxIdleConns(1)`
- **ArrearsList 分页**：ListUnpaid 支持 offset/limit
- **添加数据库索引**：contracts.status、payments.contract_id、payments.voided
- **Dockerfile 非 root 用户**：添加 appuser 运行

## Capabilities

### New Capabilities
- `cors-middleware`: CORS 跨域请求控制中间件，支持环境变量配置允许的 origin

### Modified Capabilities
- `repo-security-hardening`: BackupHandler 路径注入修复、路径泄露移除、连接泄漏修复
- `pii-masking`: 租户 Get 接口的身份证号脱敏

## Impact

- **后端 Handler**: backup.go、auth.go、payment.go、tenant.go、arrears.go
- **后端 Config**: config.go 新增 CORSOrigins 字段
- **后端 DI/路由**: main.go 新增 CORS 中间件、自定义 Recovery
- **后端 Repository**: repos.go（ArrearsList 分页）、sqlite/setup.go（连接池）
- **Domain Model**: contract.go、payment.go、receipt.go（索引 tag）
- **前端**: api/index.ts（401 处理）、ArrearsList.vue（分页适配）
- **部署**: Dockerfile（非 root 用户）、go.mod（gin-contrib/cors 依赖）
