## MODIFIED Requirements

### Requirement: 租户详情 API SHALL 对身份证号字段进行脱敏

租户详情接口（GET /api/tenants/:id）返回的身份证号字段 SHALL 将中间部分替换为星号，仅保留前 4 位和后 4 位，与列表接口行为一致。

#### Scenario: 正常身份证号脱敏
- **WHEN** 租户的身份证号为 `310101199001011234`（18 位）
- **THEN** 详情接口返回的身份证号 SHALL 为 `3101**********1234`

#### Scenario: 短身份证号不脱敏
- **WHEN** 租户的身份证号长度不超过 8 个字符
- **THEN** 详情接口返回原值不做脱敏

#### Scenario: 空身份证号
- **WHEN** 租户的身份证号为空字符串
- **THEN** 详情接口返回空字符串
