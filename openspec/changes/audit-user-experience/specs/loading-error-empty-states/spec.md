## ADDED Requirements

### Requirement: 列表页面加载状态

所有列表页面在数据加载期间 SHALL 显示加载中指示器。

#### Scenario: 页面初始加载
- **WHEN** 用户导航到列表页面（资产、租户、合同、收据本、收款）
- **THEN** 页面显示"加载中..."文字，不显示空表格

#### Scenario: 搜索触发加载
- **WHEN** 用户在搜索框输入关键词触发新请求
- **THEN** 页面显示加载中指示器，直到请求完成

### Requirement: 列表页面错误状态

所有列表页面在 API 请求失败时 SHALL 显示错误提示和重试按钮。

#### Scenario: API 请求失败
- **WHEN** 列表数据 API 返回错误（5xx、网络错误）
- **THEN** 页面显示错误提示信息和"重试"按钮

#### Scenario: 重试操作
- **WHEN** 用户点击"重试"按钮
- **THEN** 重新发起 API 请求，页面恢复加载中状态

### Requirement: 列表页面空状态

所有列表页面在数据为空时 SHALL 显示友好的空状态提示。

#### Scenario: 无数据
- **WHEN** API 返回空数组（total === 0）
- **THEN** 页面显示空状态提示（如"暂无数据"），不显示空表格头

#### Scenario: 搜索无结果
- **WHEN** 用户搜索后返回空结果
- **THEN** 页面显示"未找到匹配结果"提示

### Requirement: 搜索清除按钮

已有搜索输入框 SHALL 提供清除按钮。

#### Scenario: 清除搜索
- **WHEN** 用户点击搜索框内的 X 清除按钮
- **THEN** 搜索词清空，列表恢复为无搜索状态
