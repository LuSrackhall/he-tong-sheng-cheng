## ADDED Requirements

### Requirement: toChineseAmount 前端函数
系统 SHALL 提供 `toChineseAmount(num: number): string` 纯函数，将数字金额转为中文大写形式。函数 MUST 处理整数和小数（角/分），MUST 将 0 转为 "零元整"，MUST 将 null/undefined 转为空字符串。

#### Scenario: 整数金额转换
- **WHEN** 调用 toChineseAmount(12000)
- **THEN** 返回 "壹万贰仟元整"

#### Scenario: 小数金额转换
- **WHEN** 调用 toChineseAmount(12345.50)
- **THEN** 返回 "壹万贰仟叁佰肆拾伍元伍角"

#### Scenario: 带分的金额
- **WHEN** 调用 toChineseAmount(100.05)
- **THEN** 返回 "壹佰元零伍分"

#### Scenario: 零值
- **WHEN** 调用 toChineseAmount(0)
- **THEN** 返回 "零元整"

#### Scenario: 空值
- **WHEN** 调用 toChineseAmount(null) 或 toChineseAmount(undefined)
- **THEN** 返回 ""

### Requirement: SYSTEM_AUTO_FIELDS 注册 CN 字段
系统 SHALL 在 SYSTEM_AUTO_FIELDS 集合中注册 5 个大写字段：monthlyRentCN, yearlyRentCN, totalReceivableCN, totalReceivedCN, depositCN。这些字段 MUST 被 classifyField 识别为 'system-auto' 类型。

#### Scenario: CN 字段分类
- **WHEN** classifyField("monthlyRentCN") 被调用
- **THEN** 返回 "system-auto"

### Requirement: getSystemAutoValue 处理 CN 字段
系统 SHALL 在 getSystemAutoValue 中处理 5 个 CN 字段，调用 toChineseAmount 转换对应金额字段的值。

#### Scenario: 月租金大写值
- **WHEN** contractMonthlyRent 为 5000 且调用 getSystemAutoValue("monthlyRentCN")
- **THEN** 返回 "伍仟元整"

### Requirement: UI 大写金额展示
系统 SHALL 在合同详情表单的金额字段下方自动显示中文大写金额提示，样式为灰色小字。此展示 MUST 不依赖模板 activeFieldsList 配置，MUST 在所有金额字段（monthlyRent, yearlyRent, totalReceivable, deposit）下方自动出现。

#### Scenario: 月租金下方显示大写
- **WHEN** 用户在月租金输入框输入 5000
- **THEN** 月租金输入框下方显示灰色小字 "伍仟元整"

### Requirement: 后端 CN 占位符注册
系统 SHALL 在 buildReplaceValues 中注册 5 个 CN 占位符（monthlyRentCN, yearlyRentCN, totalReceivableCN, totalReceivedCN, depositCN），使 .docx 模板中可使用 ${monthlyRentCN} 等占位符。

#### Scenario: 模板占位符替换
- **WHEN** 合同月租金为 5000 且 .docx 模板中包含 ${monthlyRentCN}
- **THEN** 渲染后替换为 "伍仟元整"
