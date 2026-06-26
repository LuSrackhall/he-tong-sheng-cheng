## ADDED Requirements

### Requirement: Modal 无障碍属性

所有模态对话框 SHALL 包含基础无障碍属性。

#### Scenario: Modal 包含 ARIA 属性
- **WHEN** 任何模态对话框（详情弹窗、编辑弹窗、确认弹窗）被渲染
- **THEN** 对话框容器包含 role="dialog"、aria-modal="true"、aria-labelledby 属性

#### Scenario: aria-labelledby 指向标题
- **WHEN** Modal 包含标题元素
- **THEN** aria-labelledby 的值等于标题元素的 id

### Requirement: 状态标签中文显示

所有合同状态标签 SHALL 显示中文而非英文。

#### Scenario: 合同列表状态标签
- **WHEN** 用户查看合同列表或合同详情
- **THEN** 状态显示为"生效中"/"已结清"/"欠费"/"已到期"，而非"active"/"paidup"/"arrears"/"expired"

#### Scenario: 新建合同预览状态
- **WHEN** 用户在新建合同第 4 步预览合同信息
- **THEN** 状态标签显示中文

### Requirement: 动态页面标题

每个页面 SHALL 设置对应的 document.title。

#### Scenario: 页面导航标题更新
- **WHEN** 用户导航到不同页面
- **THEN** 浏览器标签页标题更新为"页面名 - 租赁管家"

#### Scenario: 默认标题
- **WHEN** 页面未配置 meta.title
- **THEN** 浏览器标签页标题显示"租赁管家"
