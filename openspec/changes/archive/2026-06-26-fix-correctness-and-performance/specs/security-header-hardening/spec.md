## ADDED Requirements

### Requirement: Content-Disposition 头编码
模板文件下载时，`Content-Disposition` 头中的文件名 SHALL 使用 RFC 5987 编码（`filename*=UTF-8''<escaped>`），同时 SHALL 剥离 `\r`、`\n`、`"` 字符防止 HTTP 头注入。

#### Scenario: 下载模板名含中文字符
- **WHEN** 模板名为 "租赁合同模板"
- **THEN** Content-Disposition 头中文件名使用 URL 编码的 UTF-8 格式

#### Scenario: 下载模板名含特殊字符
- **WHEN** 模板名为 `test"inject\r\n`
- **THEN** Content-Disposition 头中引号和换行符被剥离，不存在注入风险

### Requirement: 支付错误消息脱敏
`VoidPayment` 接口返回给客户端的错误消息 SHALL 区分业务错误和系统错误。业务错误（如"已撤销"）SHALL 保留原始消息。系统错误（数据库异常等）SHALL 返回通用消息"操作失败，请稍后重试"，同时 SHALL 在服务端日志中记录原始错误。

#### Scenario: 撤销已被撤销的收款
- **WHEN** 客户端撤销一个已经撤销的收款记录
- **THEN** 系统返回 400 错误，消息为"该收款记录已被撤销"

#### Scenario: 撤销收款时数据库异常
- **WHEN** 撤销收款过程中发生数据库错误
- **THEN** 系统返回 400 错误，消息为"操作失败，请稍后重试"，不泄露数据库内部信息
