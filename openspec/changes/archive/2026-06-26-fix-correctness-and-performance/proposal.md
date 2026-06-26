## Why

代码审查发现多个后端正确性和安全问题：输入校验缺失（角色白名单、收据本页数、合同日期）、信息泄露（支付错误消息直返客户端）、HTTP 头注入风险（Content-Disposition）、收据创建非原子操作。同时存在性能隐患：SQLite 无 WAL 模式导致并发锁争、PostgreSQL 无连接池配置、PDF 模板每次重复解析、合同重叠检测全表加载、分页无上限、静态资源无缓存头。这些修复提升系统安全性和并发性能，降低生产环境风险。

## What Changes

**正确性修复（7 项）：**
- 收据创建操作包装在数据库事务中，防止并发序号冲突
- Content-Disposition 使用 RFC 5987 编码，防止 HTTP 头注入
- CreateUser 添加角色白名单校验（仅允许 admin/operator）
- 支付撤销错误消息区分业务错误和系统错误，系统错误不返回客户端
- 合同更新日期解析失败返回 400 错误，不再静默忽略
- 收据本创建校验 TotalPages > 0
- 分页 limit 统一 clamp 到 100，offset 非负校验

**性能优化（7 项）：**
- SQLite 启用 WAL 模式 + foreign_keys + busy_timeout + synchronous=NORMAL
- PostgreSQL 设置连接池参数（MaxOpenConns=25, MaxIdleConns=10, ConnMaxLifetime=5min）
- PDF 收据和模板预览模板使用 sync.Once 预解析
- 合同重叠检测新增 CheckOverlap 接口方法，用 SQL WHERE 替代全表加载
- 分页参数统一 clamp 上限为 100
- 静态资源添加 Cache-Control 头（带哈希资源长期缓存，index.html 协商缓存）

## Capabilities

### New Capabilities
- `input-validation-hardening`: 统一输入校验加固，包括角色白名单、收据本页数验证、分页参数 clamp、合同日期格式校验
- `security-header-hardening`: HTTP 安全头加固，Content-Disposition 编码、支付错误信息脱敏
- `transaction-atomicity`: 收据创建操作事务原子性保障
- `db-performance-tuning`: 数据库性能调优，SQLite WAL 模式、PostgreSQL 连接池配置
- `template-caching`: PDF 模板编译缓存
- `overlap-detection-optimization`: 合同重叠检测 SQL 优化
- `static-asset-caching`: 静态资源 Cache-Control 策略

### Modified Capabilities

（无现有 spec 需要修改）

## Impact

**受影响代码：**
- `internal/transport/handler/` — print.go, template.go, auth.go, payment.go, contract.go, receiptbook.go, asset.go, tenant.go
- `internal/transport/middleware/spa.go` — Cache-Control 头
- `internal/repository/sqlite/setup.go` — PRAGMA 配置
- `internal/repository/postgres/setup.go` — 连接池配置
- `internal/repository/sqlite/contract.go` + `postgres/contract.go` — 新增 CheckOverlap
- `internal/domain/repo.go` — ContractRepo 接口新增方法
- `internal/pdf/receipt.go` + `contract.go` — 模板缓存

**API 影响：** 无破坏性变更。CreateUser 对非法角色返回 400（之前静默接受）；合同更新日期格式错误返回 400（之前静默忽略）；分页超限自动 clamp（之前无限制）。均为安全加固行为，前端无需修改。
