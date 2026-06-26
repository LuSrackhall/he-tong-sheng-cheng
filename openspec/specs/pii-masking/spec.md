## ADDED Requirements

### Requirement: 租户列表 API SHALL 对身份证号字段进行脱敏

租户列表接口（GET /api/tenants）返回的身份证号字段 SHALL 将中间部分替换为星号，仅保留前 4 位和后 4 位。

#### Scenario: 正常身份证号脱敏
- **WHEN** 租户的身份证号为 `310101199001011234`（18 位）
- **THEN** 列表接口返回的身份证号 SHALL 为 `3101**********1234`

#### Scenario: 短身份证号不脱敏
- **WHEN** 租户的身份证号长度不超过 6 个字符
- **THEN** 列表接口返回原值不做脱敏

#### Scenario: 空身份证号
- **WHEN** 租户的身份证号为空字符串
- **THEN** 列表接口返回空字符串

#### Scenario: 详情接口返回完整值
- **WHEN** 通过 GET /api/tenants/:id 获取单个租户详情
- **THEN** 身份证号 SHALL 返回完整原始值（供编辑使用）
