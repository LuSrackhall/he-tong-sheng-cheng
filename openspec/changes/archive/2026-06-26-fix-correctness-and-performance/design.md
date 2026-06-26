## Context

资产租赁系统后端存在多处输入校验缺失和安全隐患，需要系统性加固。同时存在数据库性能瓶颈和前端资源缓存缺失。

当前状态：
- 分页参数无上限校验，可被滥用
- 用户角色无白名单，可创建任意角色
- 收据本页数无正数验证
- 合同日期解析失败静默忽略
- Content-Disposition 未编码，存在头注入风险
- 支付错误消息直接泄露数据库内部信息
- SQLite 无 WAL 模式，并发写入频繁锁争
- PostgreSQL 无连接池配置
- PDF 模板每次请求重新解析
- 合同重叠检测全表加载后内存过滤
- 静态资源无 Cache-Control 头

## Goals / Non-Goals

**Goals:**
- 修复所有已知的输入校验漏洞和安全问题
- 提升数据库并发性能和查询效率
- 添加静态资源缓存策略
- 所有修复向后兼容，不改变 API 契约

**Non-Goals:**
- 不修改浮点精度（float64 到 decimal，影响面过大）
- 不修复代码库中不存在的问题（前端 base64/payMethod/Payment.vue、UpdateUser handler）
- 不重构整体架构或引入新框架

## Decisions

### 输入校验架构

选择在 handler 层（transport 层）集中校验，而非在 repository 层。原因：
- handler 是 API 入口，尽早拒绝无效输入
- 避免校验逻辑散落在各层
- 与现有 `binding:"required"` 模式一致

抽取 `parsePagination` 工具函数到 `internal/transport/handler/pagination.go`，所有列表接口统一调用。

### 事务边界策略

收据创建（print.go）的 `GetActive` → `AllocateSequence` → `Create` 操作包装在 `h.db.Transaction` 回调中。`AllocateSequence` 本身的 UPDATE + 1 操作是数据库级原子的，但与后续 Create 之间缺少事务保护，可能导致序号空洞。

### SQLite PRAGMA 配置

通过 `gorm.Open` 后 `db.Exec()` 设置 WAL 模式、foreign_keys、busy_timeout 和 synchronous。在 AutoMigrate 之前执行，确保迁移操作也受益于 WAL 模式。

### PostgreSQL 连接池

在 `postgres.Setup` 函数中通过 `db.DB()` 获取底层 `*sql.DB` 设置连接池参数。使用硬编码值而非配置项，简化实现。后续如需调优可从环境变量读取。

### PDF 模板缓存

使用包级 `sync.Once` + `*template.Template` 变量。`receipt.go` 和 `contract.go` 各自独立缓存。选择 `sync.Once` 而非 `init()` 函数，因为 `template.Must` + `init()` 在模板语法错误时会导致包初始化 panic，而 `sync.Once` 可以返回 error。

### 合同重叠检测

在 `ContractRepo` 接口新增 `CheckOverlap(assetID, tenantID uint, start, end time.Time) (bool, error)` 方法。SQL 条件：`asset_id = ? AND tenant_id = ? AND status IN ('active','arrears') AND start_date < ? AND end_date > ?`。sqlite 和 postgres 两套仓库都需要实现。

### Content-Disposition 编码

使用 `url.PathEscape` 编码文件名，输出格式为 `attachment; filename="ascii-safe"; filename*=UTF-8''<escaped>`。同时剥离 `\r`、`\n`、`"` 字符。

### 支付错误脱敏

对 `VoidPayment` handler 中 `err.Error()` 的处理：已知业务错误（如"该收款记录已被撤销"）通过 `fmt.Errorf` 创建，前缀匹配可识别。系统错误（数据库异常等）返回通用消息"操作失败，请稍后重试"，同时 `log.Printf` 记录原始错误。

### 静态资源缓存策略

在 `spa.go` 中间件中，根据路径设置不同 Cache-Control：
- `index.html` → `no-cache`（协商缓存，确保能获取最新版本）
- `assets/` 路径（Vite 构建产物带内容哈希）→ `public, max-age=31536000, immutable`
- 其他静态文件 → `public, max-age=3600`

## Risks / Trade-offs

- **[R1] AllocateSequence 两步操作** → Update + First 非真正原子。本批次将调用端包装在事务中减少风险，但底层实现的 RETURNING 优化留待后续。
- **[R2] 合同重叠检测接口变更** → 新增接口方法需两个 repository 都实现。如有其他分支修改同一接口，合并时可能冲突。
- **[R3] SQLite WAL 模式** → WAL 在 NFS 等网络文件系统上有已知问题。本系统为单机部署，不受影响。
- **[R4] 支付错误分类** → 依赖 error message 前缀匹配区分业务/系统错误，不够优雅但实用。后续可引入自定义 error type。
