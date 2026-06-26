## Context

资产租赁与催缴管理系统（Go 后端 Gin + GORM，Vue 3 前端）存在 11 个安全漏洞，涵盖认证安全、数据安全、基础设施安全三个层面。当前项目支持 SQLite 和 PostgreSQL 双数据库，JWT 认证，管理员/操作员双角色。

安全审计发现的问题：
- 默认管理员密码 admin123 硬编码
- JWT 签名算法未显式校验，存在 alg:none 攻击风险
- 登录接口无暴力破解防护
- 租户身份证号 API 全量返回无脱敏
- 无请求体大小限制
- PostgreSQL sslmode=disable 硬编码
- 备份恢复确认通过 query string（CSRF 风险）
- 错误消息泄露内部信息（数据库表名、列名等）
- SQLite 未配置 WAL/foreign_keys/busy_timeout
- bcrypt 错误被忽略
- 存在 dead code

## Goals / Non-Goals

**Goals:**
- 消除所有 11 个已识别的安全漏洞
- 不改变现有业务逻辑和用户工作流
- 保持向后兼容（默认密码策略除外，需文档更新）
- 前端同步修改以适配后端安全变更

**Non-Goals:**
- 不做错误处理架构重构（BusinessError 等），留作后续 change
- 不引入新的外部依赖（rate limiter 使用内存 map 实现）
- 不做前端 UI 改版
- 不修改现有 API 接口签名（除 backup restore 的确认方式）

## Decisions

### 决策 1：默认管理员密码策略

**选择**：未设置 ADMIN_PASSWORD 时启动失败（log.Fatalf）

**理由**：与现有 JWT_SECRET 策略保持一致（config.go:56 已有先例）。强制部署者设置密码，消除默认密码风险。

**替代方案**：
- 首次登录强制修改密码：增加前端改动和复杂度，不如启动时直接拦截
- 仅生产环境强制：难以可靠区分环境，可能在生产环境误用默认密码

### 决策 2：Rate Limiter 实现

**选择**：内存 map[string]*loginAttempt + sync.Mutex，5 分钟窗口内最多 5 次失败

**理由**：单实例内部管理系统，非高并发场景。无需新依赖，简单可靠。重启清零可接受（登录保护本身就是临时状态）。

**关键设计**：
- key 粒度：IP 地址（通过 c.ClientIP() 获取）
- 后台 goroutine 每 10 分钟清理过期条目，防止内存泄漏
- 触发限流时返回 HTTP 429 状态码
- 新建 `internal/security/ratelimit.go` 封装

**替代方案**：
- golang.org/x/time/rate：令牌桶算法更成熟，但对本场景过度工程化，且需引入新依赖

### 决策 3：PII 脱敏位置

**选择**：Handler 层脱敏，仅对 List 接口的响应做 mask

**理由**：管理员编辑租户时需要看到完整身份证号。在 handler 层可以灵活控制哪些接口脱敏。

**关键设计**：
- 定义统一的 `maskIDCard(idCard string) string` 辅助函数
- List 接口：始终返回脱敏值
- Get 接口：返回完整值（供编辑使用，需要认证）
- Create/Update 响应：不做脱敏（用户刚输入的值）

**替代方案**：
- Repository 层脱敏：全局一致但编辑场景变复杂，需要额外的 "raw" 查询路径

### 决策 4：错误消息处理

**选择**：逐个替换 `err.Error()` 直接返回的地方为通用消息

**原则**：仅脱敏 5xx 系统级错误，保留 4xx 业务校验错误的具体消息。

**替代方案**：
- 定义 BusinessError 类型：属于错误处理架构重构，超出本次安全加固范围

### 决策 5：SQLite PRAGMA 配置

**选择**：在 `sqlite.Setup()` 中 `gorm.Open()` 后立即通过 `db.Exec()` 执行

**PRAGMA 列表**：
```sql
PRAGMA journal_mode=WAL;      -- 并发读写性能提升
PRAGMA foreign_keys=ON;        -- 启用外键约束
PRAGMA busy_timeout=5000;      -- 锁等待 5 秒，减少 "database is locked" 错误
```

**理由**：初始化阶段一次性配置，db.Exec() 确保在已有连接上执行生效。

### 决策 6：PostgreSQL SSL 配置

**选择**：新增 `DB_SSLMODE` 环境变量，默认 `disable`（向后兼容）

**理由**：通过 config.go 加载，传递到 postgres.Setup()。文档建议生产环境使用 `require`。

## Risks / Trade-offs

- **[Rate limiter 内存泄漏]** → 缓解：后台 goroutine 每 10 分钟清理过期条目
- **[PRAGMA 在连接池中不生效]** → 缓解：使用 db.Exec() 确保在已有连接上执行；SQLite 连接池通常只有一个连接
- **[5xx 错误脱敏后调试困难]** → 缓解：服务端日志（gin.Logger）保留完整错误信息，仅 API 响应脱敏
- **[默认密码策略影响开发体验]** → 缓解：文档说明开发时设置环境变量 `ADMIN_PASSWORD=dev123`
- **[前端备份恢复接口变更]** → 缓解：同步修改前端 api/index.ts，confirmed 从 URL 移到 FormData
