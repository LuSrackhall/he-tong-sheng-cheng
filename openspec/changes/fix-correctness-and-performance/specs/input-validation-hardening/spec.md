## ADDED Requirements

### Requirement: 分页参数统一校验
系统 SHALL 对所有列表接口的 `offset` 和 `limit` 查询参数进行校验：`offset` SHALL 非负（负值自动修正为 0），`limit` SHALL 在 1 到 100 之间（超出范围自动 clamp）。非数字输入 SHALL 回退为默认值（offset=0, limit=20）。

#### Scenario: limit 超过上限被 clamp
- **WHEN** 客户端请求 `GET /api/assets?limit=999`
- **THEN** 系统返回最多 100 条记录

#### Scenario: limit 为负数回退默认值
- **WHEN** 客户端请求 `GET /api/tenants?limit=-5`
- **THEN** 系统使用 limit=20 返回结果

#### Scenario: offset 为负数修正为 0
- **WHEN** 客户端请求 `GET /api/contracts?offset=-10`
- **THEN** 系统使用 offset=0 返回结果

#### Scenario: 非数字参数回退默认值
- **WHEN** 客户端请求 `GET /api/assets?limit=abc&offset=xyz`
- **THEN** 系统使用 offset=0, limit=20 返回结果

### Requirement: 用户角色白名单校验
`CreateUser` 接口 SHALL 仅接受 `admin` 和 `operator` 两种角色值。空值 SHALL 默认为 `operator`。其他任何值 SHALL 返回 400 错误。

#### Scenario: 创建用户时指定有效角色
- **WHEN** 客户端提交 `{"username":"test","password":"pass","role":"admin"}`
- **THEN** 系统创建角色为 admin 的用户

#### Scenario: 创建用户时角色为空
- **WHEN** 客户端提交 `{"username":"test","password":"pass","role":""}`
- **THEN** 系统创建角色为 operator 的用户

#### Scenario: 创建用户时角色非法
- **WHEN** 客户端提交 `{"username":"test","password":"pass","role":"superadmin"}`
- **THEN** 系统返回 400 错误，消息提示角色不合法

### Requirement: 收据本 TotalPages 正数校验
创建收据本时 `totalPages` SHALL 大于 0。值为 0 或负数 SHALL 返回 400 错误。

#### Scenario: 创建收据本时 totalPages 为 0
- **WHEN** 客户端提交 `{"prefix":"A","totalPages":0}`
- **THEN** 系统返回 400 错误

#### Scenario: 创建收据本时 totalPages 为负数
- **WHEN** 客户端提交 `{"prefix":"A","totalPages":-10}`
- **THEN** 系统返回 400 错误

#### Scenario: 创建收据本时 totalPages 为正数
- **WHEN** 客户端提交 `{"prefix":"A","totalPages":100}`
- **THEN** 系统成功创建收据本

### Requirement: 合同日期格式校验
更新合同时，如果提供了 `startDate` 或 `endDate` 且格式不合法（非 `YYYY-MM-DD` 格式），系统 SHALL 返回 400 错误，而非静默忽略。

#### Scenario: 更新合同日期格式正确
- **WHEN** 客户端提交 `{"startDate":"2024-06-01"}`
- **THEN** 系统更新合同开始日期

#### Scenario: 更新合同日期格式错误
- **WHEN** 客户端提交 `{"startDate":"not-a-date"}`
- **THEN** 系统返回 400 错误，提示日期格式不正确

#### Scenario: 更新合同日期字段为空
- **WHEN** 客户端提交 `{"startDate":""}`
- **THEN** 系统保留原日期不变
