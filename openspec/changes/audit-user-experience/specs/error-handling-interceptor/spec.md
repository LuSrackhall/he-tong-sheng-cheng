## ADDED Requirements

### Requirement: Axios 拦截器全局错误处理

Axios 响应拦截器 SHALL 处理 5xx 和网络错误。

#### Scenario: 服务器错误
- **WHEN** API 返回 5xx 状态码
- **THEN** 显示 toast 错误提示"服务器错误，请稍后重试"

#### Scenario: 网络错误
- **WHEN** 请求因网络问题失败（无响应）
- **THEN** 显示 toast 错误提示"网络连接失败，请检查网络"

#### Scenario: 错误继续传播
- **WHEN** 拦截器处理完错误
- **THEN** 错误继续抛出（reject），供 view 级 try/catch 处理上下文 UI

### Requirement: 列表页面本地错误处理

每个列表页面 SHALL 用 try/catch 包裹 API 调用，设置本地 error 状态。

#### Scenario: 加载失败显示内联错误
- **WHEN** 列表页面的 fetch 函数抛出异常
- **THEN** 页面设置 error 状态，显示内联错误 UI 和重试按钮（拦截器同时显示 toast）
