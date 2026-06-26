## Why

安全审计发现项目存在 11 个安全漏洞，涵盖认证安全（默认密码、JWT 算法校验、暴力破解防护）、数据安全（PII 泄露、请求体限制、CSRF、错误消息泄露）、基础设施安全（数据库配置、错误处理、代码质量）三个层面。这些漏洞可能导致未授权访问、数据泄露和服务拒绝，需要在下一次发布前修复。

## What Changes

1. **默认管理员密码强制配置** — 未设置 `ADMIN_PASSWORD` 环境变量时服务拒绝启动（**BREAKING**：部署流程需更新）
2. **JWT 签名算法校验** — `ParseToken` 中显式限制为 HMAC 系列（HS256/HS384/HS512），防止 alg:none 攻击
3. **登录暴力破解防护** — 新增基于 IP 的 rate limiter（内存 map + mutex，5 分钟 5 次限制）
4. **租户身份证号脱敏** — List 接口返回时身份证号中间部分替换为 `****`
5. **请求体大小限制** — 设置 `MaxMultipartMemory = 10MB`
6. **PostgreSQL SSL 可配置** — 新增 `DB_SSLMODE` 环境变量，默认 `disable`（向后兼容）
7. **备份恢复 CSRF 防护** — `confirmed` 参数从 URL query string 改为 POST body
8. **错误消息脱敏** — 5xx 错误返回通用消息，不泄露数据库内部信息
9. **SQLite PRAGMA 配置** — 连接后执行 WAL、foreign_keys、busy_timeout
10. **bcrypt 错误处理** — 检查 `GenerateFromPassword` 返回的错误
11. **Dead code 清理** — 删除 `var _ *gorm.DB = nil`

## Capabilities

### New Capabilities
- `login-rate-limiting`: 基于 IP 的登录暴力破解防护，内存 map + mutex 实现
- `pii-masking`: 租户身份证号 API 响应脱敏

### Modified Capabilities
- `repo-security-hardening`: 扩展安全配置范围（ADMIN_PASSWORD 强制、DB_SSLMODE 可配置、JWT 算法校验）

## Impact

- **部署流程**：必须设置 `ADMIN_PASSWORD` 环境变量，否则服务无法启动
- **前端 API 调用**：备份恢复接口 `confirmed` 参数需从 URL 移到 body；登录接口需处理 429 状态码
- **数据库**：SQLite 连接将启用 WAL 模式和外键约束（行为变化）
- **新增文件**：`internal/security/ratelimit.go`
- **无新外部依赖**：rate limiter 使用标准库实现
