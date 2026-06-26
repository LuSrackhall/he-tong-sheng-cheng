## MODIFIED Requirements

### Requirement: Gitignore shall exclude common sensitive patterns

（此要求未变更，保持原样。）

#### Scenario: Developer creates .env file
- **WHEN** a developer creates `.env` or `.env.local` in the project root
- **THEN** git SHALL NOT track the file and `git status` SHALL NOT show it as untracked

#### Scenario: macOS creates .DS_Store
- **WHEN** macOS generates `.DS_Store` files in any directory
- **THEN** git SHALL NOT track them

#### Scenario: Python generates cache files
- **WHEN** Python creates `__pycache__/` directories or `*.pyc` files
- **THEN** git SHALL NOT track them

#### Scenario: Log files are generated
- **WHEN** the application or tools generate `*.log` files
- **THEN** git SHALL NOT track them

### Requirement: Build artifacts shall not be tracked

（此要求未变更，保持原样。）

#### Scenario: dist files removed from index
- **WHEN** `git rm --cached -r cmd/server/dist/` is executed and committed
- **THEN** `git ls-files cmd/server/dist/` SHALL return empty output
- **AND** the local `cmd/server/dist/` directory SHALL remain intact on disk

#### Scenario: Frontend build does not re-add dist
- **WHEN** `npm run build` is run (which outputs to `cmd/server/dist/`)
- **THEN** `git status` SHALL NOT show any files under `cmd/server/dist/` as untracked or modified

### Requirement: JWT secret shall require explicit configuration

（此要求未变更，保持原样。）

#### Scenario: JWT_SECRET is set
- **WHEN** the server starts with `JWT_SECRET=my-production-secret`
- **THEN** the server SHALL start normally and use that value as the JWT signing key

#### Scenario: JWT_SECRET is unset
- **WHEN** the server starts without `JWT_SECRET` environment variable
- **THEN** the server SHALL log `FATAL: JWT_SECRET environment variable is required` and exit with code 1

#### Scenario: JWT_SECRET is empty string
- **WHEN** the server starts with `JWT_SECRET=""`
- **THEN** the server SHALL log `FATAL: JWT_SECRET environment variable is required` and exit with code 1

## ADDED Requirements

### Requirement: Admin password shall require explicit configuration

服务 SHALL 要求设置 `ADMIN_PASSWORD` 环境变量。未设置或为空时，服务 SHALL 记录致命错误并退出。

#### Scenario: ADMIN_PASSWORD is set
- **WHEN** 服务器启动时设置了 `ADMIN_PASSWORD=securepass123`
- **THEN** 服务器 SHALL 正常启动并使用该值创建默认管理员账户

#### Scenario: ADMIN_PASSWORD is unset
- **WHEN** 服务器启动时未设置 `ADMIN_PASSWORD` 环境变量
- **THEN** 服务器 SHALL 记录致命错误并以代码 1 退出

### Requirement: JWT token parsing SHALL verify signing algorithm

JWT 令牌解析 SHALL 显式验证签名方法为 HMAC 系列（HS256/HS384/HS512），拒绝所有其他签名方法。

#### Scenario: Valid HS256 token accepted
- **WHEN** 收到使用 HS256 签名的 JWT 令牌
- **THEN** 令牌 SHALL 被成功解析

#### Scenario: alg:none attack rejected
- **WHEN** 收到签名方法为 `none` 的 JWT 令牌
- **THEN** 解析 SHALL 失败并返回签名无效错误

#### Scenario: RSA-signed token rejected
- **WHEN** 收到使用 RS256 签名的 JWT 令牌
- **THEN** 解析 SHALL 失败并返回签名方法不匹配错误

### Requirement: PostgreSQL database connection SSL mode SHALL be configurable

PostgreSQL 连接的 sslmode SHALL 通过 `DB_SSLMODE` 环境变量配置。未设置时默认为 `disable`（向后兼容）。

#### Scenario: DB_SSLMODE is set to require
- **WHEN** 设置 `DB_SSLMODE=require`
- **THEN** 数据库连接 SHALL 使用 `sslmode=require`

#### Scenario: DB_SSLMODE is unset
- **WHEN** 未设置 `DB_SSLMODE` 环境变量
- **THEN** 数据库连接 SHALL 使用 `sslmode=disable`

### Requirement: Backup restore SHALL read confirmation from POST body

备份恢复接口 SHALL 从 POST 请求体（multipart form）中读取 `confirmed` 字段，而非 URL query string。

#### Scenario: Confirmation in POST body
- **WHEN** POST /api/admin/restore 请求的 form body 中包含 `confirmed=true` 和备份文件
- **THEN** 恢复操作 SHALL 被执行

#### Scenario: Confirmation in query string rejected
- **WHEN** POST /api/admin/restore?confirmed=true 请求的 form body 中不包含 confirmed 字段
- **THEN** SHALL 返回 HTTP 400 错误 "恢复操作需要确认"

### Requirement: API error messages SHALL NOT leak internal information

5xx 系统级错误 SHALL 返回通用错误消息，不包含数据库表名、列名、SQL 语句等内部信息。4xx 业务校验错误可保留具体消息。

#### Scenario: Database error masked
- **WHEN** 数据库操作返回错误（如 constraint violation）
- **THEN** API 响应 SHALL 返回通用消息（如"操作失败，请重试"），不包含 err.Error() 原始内容

#### Scenario: Business validation error preserved
- **WHEN** 业务校验失败（如"该收款记录已被撤销"）
- **THEN** API 响应 SHALL 返回具体的业务错误消息

### Requirement: SQLite SHALL be configured with security and performance PRAGMAs

SQLite 数据库连接 SHALL 在初始化时执行以下 PRAGMA：journal_mode=WAL、foreign_keys=ON、busy_timeout=5000。

#### Scenario: PRAGMAs applied on startup
- **WHEN** SQLite 模式下服务器启动
- **THEN** 数据库 SHALL 以 WAL 模式运行，外键约束 SHALL 启用，锁等待超时 SHALL 为 5 秒

### Requirement: bcrypt password hashing errors SHALL NOT be ignored

bcrypt.GenerateFromPassword 的错误 SHALL 被检查。失败时服务 SHALL 记录致命错误并退出。

#### Scenario: bcrypt hashing fails
- **WHEN** bcrypt.GenerateFromPassword 返回错误
- **THEN** 服务器 SHALL 记录致命错误并以代码 1 退出

### Requirement: Request body size SHALL be limited

服务器 SHALL 设置最大 multipart 内存为 10MB。超过限制的请求 SHALL 被拒绝。

#### Scenario: Large request rejected
- **WHEN** 收到 multipart 请求体超过 10MB
- **THEN** 服务器 SHALL 返回 HTTP 413 或类似错误
