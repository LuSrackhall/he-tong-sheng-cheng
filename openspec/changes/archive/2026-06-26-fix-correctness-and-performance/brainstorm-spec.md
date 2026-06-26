# 设计文档：正确性修复 + 性能优化

## Context

资产租赁与催缴管理系统（Go + Gin + GORM + Vue 3）在代码审查中发现了多个正确性和性能问题。本批次修复聚焦于后端安全加固、数据一致性保障和查询性能优化。

**原始问题清单调整：**

前端问题 #1（base64-b64 null 检查）、#2（payMethod 空字符串）、#3（收据撤销不同步）、#17（vendor 包过大）涉及的文件（`Payment.vue`、`useContractTemplate.js`）和依赖（`base64-b64`、`file-saver`）在当前代码库中不存在，排除。
后端问题 #6（UpdateUser 可清空密码）对应的 handler 不存在，排除。

**实际修复范围：7 个正确性问题 + 7 个性能问题**

## Goals / Non-Goals

**Goals:**
- 修复 7 个后端正确性安全问题（输入校验、注入防护、错误泄露、事务原子性）
- 修复 7 个后端性能问题（SQLite WAL、连接池、模板缓存、SQL 优化、分页限制、静态资源缓存）
- 不引入破坏性变更，所有修复向后兼容

**Non-Goals:**
- 不修改浮点精度（float64 → decimal 影响面太大，另案处理）
- 不修复不存在的前端问题
- 不重构整体架构

## Decisions

### D1: 合同重叠检测 — 新增 `CheckOverlap` 接口方法

**选择**: 在 `ContractRepo` 接口新增 `CheckOverlap(assetID, tenantID uint, start, end time.Time) (bool, error)` 方法，使用 SQL WHERE 条件精确查询。

**替代方案**: 保留 `ListActive()` + 内存过滤。放弃原因：O(N) 全表扫描。

**影响**: 需修改 `domain/repo.go` 接口、sqlite 和 postgres 两套实现、handler 调用代码。

### D2: 收据创建事务边界 — 包装在 `db.Transaction` 中

**选择**: 将 `PrintReceipt` handler 中的 `GetActive` → `AllocateSequence` → `Create` 包装在 `h.db.Transaction` 回调中。

**分析**: `AllocateSequence` 本身使用 `UPDATE ... SET current_num = current_num + 1` 是原子的，但与 `Create` 之间没有事务保护。并发请求可能同时分配序号，然后只有一个 Create 成功，另一个序号成为"空洞"。

### D3: Content-Disposition — 使用 `url.PathEscape` + `filename*` 编码

**选择**: 使用 `url.PathEscape` 对文件名进行编码，使用 RFC 5987 `filename*=UTF-8''<encoded>` 格式。同时剥离换行符和引号防止 HTTP 头注入。

### D4: 用户角色验证 — 白名单校验

**选择**: 在 `CreateUser` handler 中添加角色白名单 `["admin", "operator"]`，拒绝其他值。

### D5: 支付错误消息 — 区分业务错误和系统错误

**选择**: 对已知业务错误（如"已撤销"）保留原始消息，对系统错误返回通用消息并记录日志。

### D6: 合同日期解析 — 返回 400 错误

**选择**: 日期格式不合法时返回 `400 Bad Request`，不再静默忽略。

### D7: 收据本 TotalPages 校验 — 验证 > 0

**选择**: 在 handler 中添加 `req.TotalPages <= 0` 检查，返回 400。

### D8: 分页 limit 统一 clamp — 抽取工具函数

**选择**: 抽取 `parsePagination(c *gin.Context, defaultLimit, maxLimit int) (offset, limit int)` 工具函数，统一处理 offset 非负、limit 范围校验。asset/tenant/contract 三个 handler 调用该函数。

### D9: SQLite PRAGMA — WAL + foreign_keys + busy_timeout

**选择**: 在 `sqlite.Setup` 的 `gorm.Open` 之后添加：
```sql
PRAGMA journal_mode=WAL
PRAGMA foreign_keys=ON
PRAGMA busy_timeout=5000
PRAGMA synchronous=NORMAL
```

### D10: PostgreSQL 连接池 — 在 `Setup` 中配置

**选择**: 在 `postgres.Setup` 函数中，`gorm.Open` 成功后获取 `db.DB()` 并设置：
```go
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

### D11: PDF 模板缓存 — 包级 `sync.Once`

**选择**: 使用 `sync.Once` + 包级 `*template.Template` 变量，首次调用时解析，后续复用。`receipt.go` 和 `contract.go` 各一个。

### D12: 静态资源 Cache-Control — 区分策略

**选择**:
- `index.html` → `Cache-Control: no-cache`（协商缓存）
- `assets/` 路径下的文件 → `Cache-Control: public, max-age=31536000, immutable`（Vite 构建产物带哈希）
- 其他文件 → `Cache-Control: public, max-age=3600`

## Risks / Trade-offs

- **[R1] AllocateSequence 原子性** → 当前实现 Update + First 是两步操作。本批次暂不修改其底层实现（需要改 repository 接口签名），仅确保调用端在事务中使用。后续可单独优化为 RETURNING 子句。
- **[R2] SQLite WAL 模式** → WAL 在 NFS 等网络文件系统上有已知问题。本系统是单机部署，不影响。
- **[R3] 连接池硬编码** → 使用固定值而非配置项，简化实现。后续如需调优可从环境变量读取。
- **[R4] 合同重叠检测接口变更** → 新增接口方法，需两个 repository 都实现。如果有其他分支在修改同一接口，合并时可能冲突。
