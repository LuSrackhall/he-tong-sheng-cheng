## MODIFIED Requirements

### Requirement: BackupHandler SHALL NOT expose server file paths

BackupInfo 接口（GET /api/admin/backup/info）返回的响应 SHALL NOT 包含数据库文件的服务器内部路径。

#### Scenario: SQLite 模式下 BackupInfo 不返回路径
- **WHEN** 管理员请求 GET /api/admin/backup/info（SQLite 模式）
- **THEN** 响应 SHALL 包含 `type`、`size`、`lastModified` 字段
- **AND** 响应 SHALL NOT 包含 `path` 字段

#### Scenario: PostgreSQL 模式下 BackupInfo 不受影响
- **WHEN** 管理员请求 GET /api/admin/backup/info（PostgreSQL 模式）
- **THEN** 响应 SHALL 仅包含 `type` 和 `message` 字段

### Requirement: VACUUM INTO SHALL use safe path construction

Backup 接口（POST /api/admin/backup）执行 VACUUM INTO 时 SHALL 使用安全的路径构造方式，避免 SQL 注入风险。

#### Scenario: 正常备份生成安全 SQL
- **WHEN** 管理员请求 POST /api/admin/backup
- **THEN** VACUUM INTO 的路径 SHALL 由服务端生成（基于 backupDir + 时间戳）
- **AND** SQL 语句 SHALL 使用 `fmt.Sprintf` 构造，路径包裹在单引号中

### Requirement: Restore SHALL NOT manually close database connection

Restore 接口（POST /api/admin/restore）成功后 SHALL NOT 手动调用 `sqlDB.Close()`。数据库连接 SHALL 由优雅关停流程统一管理。

#### Scenario: 恢复成功后正常重启
- **WHEN** 管理员成功恢复数据库备份
- **THEN** 服务 SHALL 通过 shutdownFn 触发优雅关停
- **AND** 在关停前的所有在途请求 SHALL 正常完成
- **AND** 数据库连接 SHALL NOT 在 Restore handler 中被手动关闭
