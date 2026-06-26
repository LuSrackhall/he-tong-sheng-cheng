## ADDED Requirements

### Requirement: CORS 中间件 SHALL 支持环境变量配置

当设置了 `CORS_ORIGINS` 环境变量时，服务 SHALL 启用 CORS 中间件，仅允许指定的 origin 跨域请求。未设置时 SHALL 不启用 CORS（同源模式）。

#### Scenario: CORS_ORIGINS 设置了单个 origin
- **WHEN** 启动时 `CORS_ORIGINS=https://example.com`
- **THEN** 来自 `https://example.com` 的跨域请求 SHALL 获得正确的 CORS 响应头
- **AND** 来自其他 origin 的请求 SHALL NOT 获得 CORS 允许头

#### Scenario: CORS_ORIGINS 设置了多个 origin
- **WHEN** 启动时 `CORS_ORIGINS=https://a.com,https://b.com`
- **THEN** 来自 `https://a.com` 和 `https://b.com` 的请求 SHALL 获得 CORS 允许头
- **AND** 来自 `https://c.com` 的请求 SHALL NOT 获得 CORS 允许头

#### Scenario: CORS_ORIGINS 未设置
- **WHEN** 启动时未设置 `CORS_ORIGINS` 环境变量
- **THEN** 服务 SHALL NOT 添加 CORS 相关响应头
- **AND** 跨域请求 SHALL 被浏览器拦截（同源策略）

#### Scenario: CORS_ORIGINS 设为通配符
- **WHEN** 启动时 `CORS_ORIGINS=*`
- **THEN** 所有来源的请求 SHALL 获得 CORS 允许头
