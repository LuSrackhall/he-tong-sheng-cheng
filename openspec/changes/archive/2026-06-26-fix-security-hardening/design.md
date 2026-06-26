## Context

资产租赁管理系统（Go Gin + GORM + Vue 3）存在 11 个安全漏洞。项目支持 SQLite/PostgreSQL 双数据库，JWT 认证，管理员/操作员双角色。本次修复涉及后端认证、数据处理、数据库初始化三个层面，以及前端的少量适配修改。

## Goals / Non-Goals

**Goals:**
- 消除 11 个已识别安全漏洞
- 保持现有业务逻辑不变
- 前端最小化适配修改

**Non-Goals:**
- 不做错误处理架构重构
- 不引入新外部依赖
- 不改 API 接口签名（除 backup restore 确认方式）

## Decisions

### 认证安全层

**JWT 算法校验（middleware/auth.go:42）**

在 `ParseToken` 的 keyFunc 回调中添加签名方法校验：

```go
func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
    }
    return m.Secret, nil
}
```

使用 `*jwt.SigningMethodHMAC` 而非精确匹配 `jwt.SigningMethodHS256`，允许 HS384/HS512，兼容未来升级。拒绝所有非 HMAC 方法（RSA、ECDSA、none）。

**Rate Limiter（新建 internal/security/ratelimit.go）**

```
结构：
  type LoginRateLimiter struct {
      mu      sync.Mutex
      attempts map[string]*loginAttempt
  }
  
  type loginAttempt struct {
      count    int
      firstAt  time.Time
  }

逻辑：
  - Allow(ip) 检查：firstAt + 5min < now → 重置；count >= 5 → 拒绝
  - RecordFailure(ip)：count++，首次记录 firstAt；窗口过期则重置计数
  - Reset(ip)：登录成功后清除记录
  - Cleanup()：每 10 分钟删除 firstAt + 5min < now 的条目

注入：
  - AuthHandler 新增 limiter 字段
  - Login 方法开头调用 limiter.Allow(c.ClientIP())
  - 失败时调用 limiter.RecordFailure()
  - 成功时调用 limiter.Reset()
  - main.go 启动 cleanup goroutine
```

**默认密码（sqlite/setup.go:37, postgres/setup.go:38）**

从 config.Load() 中加载 ADMIN_PASSWORD，未设置时 `log.Fatalf`。Setup 函数签名变更，接收 password 参数而非自行读取环境变量。这使得密码校验集中在 config 层。

### 数据安全层

**PII 脱敏（handler/tenant.go）**

在 handler 包中定义辅助函数：

```go
func maskIDCard(idCard string) string {
    if len(idCard) <= 8 {
        return idCard
    }
    // 保留前4位和后4位，中间替换为 *
    prefix := idCard[:4]
    suffix := idCard[len(idCard)-4:]
    return prefix + strings.Repeat("*", len(idCard)-8) + suffix
}
```

仅在 `TenantHandler.List` 中对返回的每个 tenant 的 IDCard 字段调用 mask。Get/Create/Update 不做脱敏。

**请求体大小（cmd/server/main.go）**

在 `gin.New()` 之后、路由注册之前添加：
```go
r.MaxMultipartMemory = 10 << 20 // 10MB
```

**PostgreSQL SSL（config.go + postgres/setup.go）**

config.go 新增 `DBSSLMode` 字段，从 `DB_SSLMODE` 环境变量读取，默认 `disable`。postgres.Setup 签名新增 sslmode 参数。DSN 拼接使用该参数。

**备份恢复确认（handler/backup.go:102）**

从 `c.Query("confirmed")` 改为从 multipart form 读取：
```go
confirmed := c.PostForm("confirmed")
if confirmed != "true" { ... }
```

**错误消息脱敏（handler/payment.go:158）**

`VoidPayment` 中 `err.Error()` 替换为固定消息。内部错误通过 `log.Printf` 记录完整信息。

### 基础设施安全层

**SQLite PRAGMA（sqlite/setup.go）**

在 `gorm.Open()` 之后、`AutoMigrate()` 之前执行：
```go
db.Exec("PRAGMA journal_mode=WAL")
db.Exec("PRAGMA foreign_keys=ON")
db.Exec("PRAGMA busy_timeout=5000")
```

**bcrypt 错误处理（sqlite/setup.go:41, postgres/setup.go:43）**

```go
hash, err := bcrypt.GenerateFromPassword(...)
if err != nil {
    log.Fatalf("Failed to hash admin password: %v", err)
}
```

**Dead code（sqlite/repos.go:202）**

删除 `var _ *gorm.DB = nil`。

## Risks / Trade-offs

- **[Rate limiter 内存]** → 定期清理 goroutine 缓解
- **[PRAGMA 连接池]** → SQLite 默认单连接，db.Exec 直接生效
- **[错误脱敏调试]** → gin.Logger 已记录完整错误
- **[部署流程变更]** → 文档更新说明 ADMIN_PASSWORD 要求
