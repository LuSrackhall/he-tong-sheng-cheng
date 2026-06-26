## 1. 后端正确性修复 — 输入校验

- [x] 1.1 新建 `internal/transport/handler/pagination.go`，实现 `parsePagination(c *gin.Context, defaultLimit, maxLimit int) (offset, limit int)` 工具函数，处理 offset 非负、limit 范围 clamp、非数字回退默认值
- [x] 1.2 修改 `internal/transport/handler/asset.go` List 方法，调用 `parsePagination` 替换原有解析逻辑
- [x] 1.3 修改 `internal/transport/handler/tenant.go` List 方法，调用 `parsePagination` 替换原有解析逻辑
- [x] 1.4 修改 `internal/transport/handler/contract.go` List 方法，调用 `parsePagination` 替换原有解析逻辑
- [x] 1.5 修改 `internal/transport/handler/auth.go` CreateUser 方法，添加角色白名单校验（仅允许 admin/operator）
- [x] 1.6 修改 `internal/transport/handler/receiptbook.go` Create 方法，添加 `req.TotalPages <= 0` 校验
- [x] 1.7 修改 `internal/transport/handler/contract.go` UpdateContract 方法，日期解析失败时返回 400 错误

## 2. 后端正确性修复 — 安全加固

- [x] 2.1 修改 `internal/transport/handler/template.go` DownloadTemplate 方法，Content-Disposition 使用 `url.PathEscape` + `filename*=UTF-8''<encoded>` 格式，剥离 `\r`、`\n`、`"` 字符
- [x] 2.2 修改 `internal/transport/handler/payment.go` VoidPayment 方法，区分业务错误和系统错误，系统错误返回通用消息并记录服务端日志

## 3. 后端正确性修复 — 事务原子性

- [x] 3.1 修改 `internal/transport/handler/print.go` PrintReceipt 方法，将 GetActive → AllocateSequence → Create 包装在 `h.db.Transaction` 回调中

## 4. 性能优化 — 数据库调优

- [x] 4.1 修改 `internal/repository/sqlite/setup.go` Setup 函数，在 gorm.Open 后添加 PRAGMA journal_mode=WAL、foreign_keys=ON、busy_timeout=5000、synchronous=NORMAL
- [x] 4.2 修改 `internal/repository/postgres/setup.go` Setup 函数，通过 db.DB() 设置连接池参数 MaxOpenConns=25、MaxIdleConns=10、ConnMaxLifetime=5min

## 5. 性能优化 — 模板缓存

- [x] 5.1 修改 `internal/pdf/receipt.go`，使用 sync.Once + 包级变量缓存解析后的 receipt 模板
- [x] 5.2 修改 `internal/pdf/contract.go`，使用 sync.Once + 包级变量缓存解析后的 templatePreview 模板

## 6. 性能优化 — 合同重叠检测 SQL 优化

- [x] 6.1 在 `internal/domain/repo.go` ContractRepo 接口中新增 `CheckOverlap(assetID, tenantID uint, start, end time.Time) (bool, error)` 方法
- [x] 6.2 在 `internal/repository/sqlite/contract.go` 实现 CheckOverlap 方法（SQL WHERE 条件查询）
- [x] 6.3 在 `internal/repository/postgres/contract.go` 实现 CheckOverlap 方法（SQL WHERE 条件查询）
- [x] 6.4 修改 `internal/transport/handler/contract.go` CreateContract 方法，使用 CheckOverlap 替代 ListActive + 内存循环

## 7. 性能优化 — 静态资源缓存

- [x] 7.1 修改 `internal/transport/middleware/spa.go` SPAFallbackEmbed 函数，根据路径设置 Cache-Control 头：index.html 使用 no-cache，assets/ 使用 immutable 长缓存，其他使用 1 小时缓存

---

## Post-Implementation Workflow

完成所有任务后，严格按以下顺序执行：

1. **验证**：运行 `go build ./...` 和 `go test ./... -count=1` 确认编译和测试通过
2. **提交**：按 task group 分组提交，中文 commit message
3. **通知主会话**：报告"就绪待合入"，等待调度
4. **合并**：收到主会话通知后执行 myspec-merge
