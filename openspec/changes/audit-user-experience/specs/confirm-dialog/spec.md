## ADDED Requirements

### Requirement: ConfirmDialog 组件

系统 SHALL 提供可复用的 ConfirmDialog 组件替代原生 confirm()。

#### Scenario: 显示确认对话框
- **WHEN** 代码调用 ConfirmDialog 显示（v-model:visible = true）
- **THEN** 显示模态对话框，包含标题、消息文本、确认按钮和取消按钮

#### Scenario: 确认操作
- **WHEN** 用户点击确认按钮
- **THEN** 触发 confirm 事件，对话框关闭

#### Scenario: 取消操作
- **WHEN** 用户点击取消按钮或点击遮罩层
- **THEN** 触发 cancel 事件，对话框关闭

#### Scenario: 危险操作样式
- **WHEN** variant 设置为 'danger'
- **THEN** 确认按钮显示红色警告样式

### Requirement: 替换原生 confirm() 调用

系统中所有原生 confirm() 调用 SHALL 替换为 ConfirmDialog 组件。

#### Scenario: 作废付款确认
- **WHEN** 用户在收款页面作废一笔付款
- **THEN** 显示 ConfirmDialog 确认"确定要作废这笔收款记录吗？"，而非原生 confirm

#### Scenario: 删除用户确认
- **WHEN** 管理员在用户管理页面删除用户
- **THEN** 显示 ConfirmDialog 确认，而非原生 confirm

#### Scenario: 模板删除确认
- **WHEN** 用户在设置页面删除模板
- **THEN** 显示 ConfirmDialog 确认，而非原生 confirm
