## Context

资产租赁与催缴管理系统已完成安全加固，审计发现 13 个 P0-P2 级别问题需要修复。系统基于 Go (Gin + GORM) 后端 + Vue 3 前端 + SQLite/PostgreSQL 数据库。

## Goals / Non-Goals

**Goals:**
- 修复全部 13 个审计发现的问题
- 保持向后兼容（API 行为变更仅为安全性修复）
- 所有修复通过编译、测试和构建验证

**Non-Goals:**
- 前端架构重构
- 新增业务功能
- 数据库 schema 重大变更

## Decisions

### D1: VACUUM INTO 安全构造

GORM 的参数化查询对 SQLite `VACUUM INTO` 不可靠。backup.go:69 使用 `h.db.Exec("VACUUM INTO ?", backupPath)` 可能生成错误 SQL。

**方案**: 使用 `fmt.Sprintf("VACUUM INTO '%s'", backupPath)` 直接构造。backupPath 由服务端 `fmt.Sprintf("%s/backup_%s.db", backupDir, time.Now().Format(...))` 生成，不接受用户输入，不存在注入风险。

### D2: 恢复后连接处理

backup.go Restore 方法手动调用 `sqlDB.Close()` 后，1 秒延迟内并发请求会访问已关闭连接。

**方案**: 移除 `sqlDB.Close()` 调用。优雅关停流程（main.go 的 `srv.Shutdown`）会等待所有在途请求完成后再关闭数据库。shutdownFn 触发 SIGTERM 即可。

### D3: 租户身份证号全面脱敏

tenant.go List 接口已脱敏，但 Get 接口（/api/tenants/:id）返回完整 IDCard。

**方案**: Get 接口也调用 `maskIDCard()`。前端编辑表单不需要完整身份证号（用户自己知道），编辑提交时携带原始值即可。实际上当前 Update 接口接受前端传入的 IDCard，所以前端编辑场景不受影响。

**影响**: 这是一个 BREAKING 行为变更，但属于安全修复范畴。

### D4: BackupInfo 路径隐藏

backup.go BackupInfo 返回 `info["path"] = h.dbPath`，暴露服务器内部文件路径。

**方案**: 移除 `path` 字段。前端 Settings.vue 未使用该字段，无前端影响。

### D5: CORS 中间件

当前无 CORS 控制，同源部署安全但不支持跨域场景。

**方案**:
1. config.go 新增 `CORSOrigins` 字段，来源 `CORS_ORIGINS` 环境变量，默认空
2. 仅当 CORS_ORIGINS 非空时启用 cors 中间件（`github.com/gin-contrib/cors`）
3. CORS_ORIGINS 格式：逗号分隔的 origin 列表，如 `https://example.com,https://app.example.com`
4. 中间件放在 `gin.New()` 之后、路由之前

**替代方案**: 硬编码允许所有 origin — 安全风险高，不采用。

### D6: 类型断言安全化

auth.go:221 `userID.(uint)` 和其他 handler 中的类型断言可能 panic。

**方案**: 统一提取 helper 函数 `getUintFromContext(c *gin.Context, key string) (uint, error)`，所有需要的地方调用此函数。包含 ok-check，失败返回 500。

### D7: VoidPayment 状态码

payment.go VoidPayment 系统错误返回 400。

**方案**: 已知业务错误（"该收款记录已被撤销"）保持 400；其他错误返回 500。调整判断逻辑使用 `errors.Is` 或字符串匹配。

### D8: Recovery 定制

main.go:73 `gin.Recovery()` 默认打印堆栈。

**方案**: 使用 `gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, err any) { ... })` 自定义处理，仅记录错误消息和请求路径，不输出堆栈。

### D9: 前端 401 处理

api/index.ts:18 使用 `window.location.hash = '#/login'` 跳转不重置状态。

**方案**: 改为 `localStorage.removeItem('token'); window.location.reload()`。页面刷新后 auth store 初始化为空，router guard 自动跳转到 /login。

### D10: SQLite 连接池

sqlite/setup.go 未配置连接池参数。

**方案**:
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(1)   // SQLite 单写模型
sqlDB.SetMaxIdleConns(1)
sqlDB.SetConnMaxLifetime(0) // 永不回收
```

### D11: ArrearsList 分页

arrears.go:34 `ListUnpaid()` 加载全部未付合同。

**方案**:
1. `ContractRepo.ListUnpaid` 改为 `ListUnpaidPaging(offset, limit int) ([]Contract, int64, error)`
2. arrears.go handler 使用 `parsePagination` 解析分页参数
3. 前端 ArrearsList.vue 适配 `{data, total}` 响应格式，添加分页控件

### D12: 数据库索引

常用查询字段缺少索引。

**方案**: 在 domain model GORM tag 中添加索引：
- `contract.go`: `Status string \`gorm:"index"\``
- `payment.go`: `ContractID uint \`gorm:"index"\``, `Voided bool \`gorm:"index"\``
- `receipt.go`: `PaymentID uint \`gorm:"index"\``
- GORM AutoMigrate 会自动创建索引

### D13: Dockerfile 非 root

当前以 root 运行。

**方案**:
```dockerfile
RUN adduser -D -u 1001 appuser
COPY --from=backend-builder --chown=appuser:appuser /app/server .
USER appuser
```

## Risks / Trade-offs

- **[Get 接口身份证脱敏]** → 可能影响需要完整身份证号的下游流程。缓解：Update 接口仍接受前端传入的 IDCard，用户编辑时前端已有完整值。
- **[SQLite MaxOpenConns(1)]** → 并发写入排队。缓解：SQLite 本身就是单写模型，这是正确配置。
- **[ArrearsList 分页]** → 前端需适配。缓解：使用统一的 parsePagination 模式，改动量小。
- **[CORS 默认不启用]** → 开发环境需手动配置。缓解：文档说明，开发时设置环境变量即可。
