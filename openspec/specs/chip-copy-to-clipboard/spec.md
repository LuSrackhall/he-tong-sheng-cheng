## ADDED Requirements

### Requirement: 复制字段占位符到剪贴板

系统 SHALL 在每个 field-chip 标签上提供独立的复制按钮，点击后将 `${字段名}` 格式的占位符文本复制到系统剪贴板。

#### Scenario: 点击复制按钮复制预置字段
- **WHEN** 用户点击预置字段 chip（如 tenantName）上的复制按钮
- **THEN** `${tenantName}` 被复制到剪贴板

#### Scenario: 点击复制按钮复制自定义字段
- **WHEN** 用户点击自定义字段 chip（如 customField）上的复制按钮
- **THEN** `${customField}` 被复制到剪贴板

#### Scenario: 复制不触发映射操作
- **WHEN** 用户点击复制按钮
- **THEN** 该 chip 的映射状态（已添加/未添加）不发生变化

### Requirement: 复制后即时反馈

系统 SHALL 在用户复制成功后显示 Tooltip "已复制"反馈，1.5 秒后自动消失。

#### Scenario: 复制成功显示 Tooltip
- **WHEN** 用户点击复制按钮且复制操作成功
- **THEN** 在复制按钮旁显示 "已复制" Tooltip

#### Scenario: Tooltip 自动消失
- **WHEN** Tooltip 已显示 1.5 秒
- **THEN** Tooltip 自动消失

#### Scenario: 快速连续复制多个字段
- **WHEN** 用户快速连续点击不同 chip 的复制按钮
- **THEN** 每个被点击的 chip 各自显示独立的 "已复制" Tooltip

### Requirement: 复制按钮不影响现有交互

复制按钮 MUST 不干扰现有的 chip-label 映射操作和 chip-toggle 启用/禁用操作。

#### Scenario: 点击 chip-label 仍触发映射操作
- **WHEN** 用户点击 chip-label 文本区域
- **THEN** 触发原有的添加/移除映射操作，不触发复制

#### Scenario: 点击 chip-toggle 仍触发切换操作
- **WHEN** 用户点击 chip-toggle 开关
- **THEN** 触发原有的启用/禁用切换，不触发复制
