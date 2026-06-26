## ADDED Requirements

### Requirement: 登录接口 SHALL 实施基于 IP 的暴力破解防护

登录接口 SHALL 对每个客户端 IP 地址实施失败次数限制。在 5 分钟窗口内累计失败 5 次后，该 IP 的后续登录请求 SHALL 被拒绝，直到窗口过期。

#### Scenario: 正常登录不受影响
- **WHEN** 客户端 IP 在 5 分钟窗口内登录失败次数少于 5 次
- **THEN** 登录请求 SHALL 正常处理（成功或返回"用户名或密码错误"）

#### Scenario: 超过失败次数限制
- **WHEN** 客户端 IP 在 5 分钟窗口内累计登录失败 5 次
- **THEN** 该 IP 的后续登录请求 SHALL 返回 HTTP 429 状态码，响应 body 包含 "登录尝试次数过多，请稍后再试"

#### Scenario: 窗口过期后恢复
- **WHEN** 客户端 IP 的首次失败记录已超过 5 分钟
- **THEN** 该 IP 的失败计数 SHALL 被重置，允许重新尝试登录

#### Scenario: 登录成功清除记录
- **WHEN** 客户端 IP 登录成功
- **THEN** 该 IP 的失败记录 SHALL 被清除

#### Scenario: 服务重启后计数重置
- **WHEN** 服务重启
- **THEN** 所有 IP 的失败记录 SHALL 被清除（内存存储特性）
