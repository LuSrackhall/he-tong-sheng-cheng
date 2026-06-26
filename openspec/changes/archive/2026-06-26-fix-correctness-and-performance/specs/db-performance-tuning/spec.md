## ADDED Requirements

### Requirement: SQLite WAL 模式启用
SQLite 数据库初始化时 SHALL 设置以下 PRAGMA：`journal_mode=WAL`、`foreign_keys=ON`、`busy_timeout=5000`、`synchronous=NORMAL`。这些设置 SHALL 在 AutoMigrate 之前执行。

#### Scenario: SQLite 数据库初始化
- **WHEN** 系统以 SQLite 模式启动
- **THEN** 数据库使用 WAL 日志模式，外键约束已启用，忙等待超时 5 秒，同步模式为 NORMAL

#### Scenario: 并发读写场景
- **WHEN** 多个请求同时读写 SQLite 数据库
- **THEN** 写操作不阻塞读操作，不出现 "database is locked" 错误

### Requirement: PostgreSQL 连接池配置
PostgreSQL 数据库初始化时 SHALL 配置连接池参数：`MaxOpenConns=25`、`MaxIdleConns=10`、`ConnMaxLifetime=5min`。

#### Scenario: PostgreSQL 数据库初始化
- **WHEN** 系统以 PostgreSQL 模式启动
- **THEN** 连接池最大连接数 25，最大空闲连接数 10，连接最大生命周期 5 分钟
