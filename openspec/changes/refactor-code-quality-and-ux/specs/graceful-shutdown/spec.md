## ADDED Requirements

### Requirement: 优雅关停信号处理

服务器 SHALL 监听 SIGINT 和 SIGTERM 信号，收到后触发优雅关停流程。

#### Scenario: 收到 SIGTERM 关闭
- **WHEN** 服务器进程收到 SIGTERM 信号
- **THEN** 停止接受新连接
- **THEN** 等待所有 in-flight 请求完成（最长 10 秒）
- **THEN** 关闭数据库连接
- **THEN** 进程正常退出

#### Scenario: 收到 SIGINT 关闭
- **WHEN** 服务器进程收到 SIGINT 信号（Ctrl+C）
- **THEN** 执行与 SIGTERM 相同的优雅关停流程

#### Scenario: 关闭超时
- **WHEN** in-flight 请求在 10 秒内未完成
- **THEN** 强制关闭剩余连接并退出

### Requirement: 数据库恢复后优雅退出

数据恢复操作完成后，SHALL 通过优雅关停流程退出而非 `os.Exit(0)`。

#### Scenario: 恢复数据库后退出
- **WHEN** 管理员完成数据库恢复操作
- **THEN** 返回成功响应给客户端
- **THEN** 等待短暂延迟（确保响应发送完成）
- **THEN** 触发优雅关停流程退出进程
