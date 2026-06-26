## ADDED Requirements

### Requirement: 静态资源 Cache-Control 策略
SPA 中间件 SHALL 根据请求路径设置不同的 `Cache-Control` 响应头：
- `index.html`：`Cache-Control: no-cache`（协商缓存）
- `assets/` 路径下文件（Vite 构建产物带内容哈希）：`Cache-Control: public, max-age=31536000, immutable`
- 其他静态文件：`Cache-Control: public, max-age=3600`

#### Scenario: 请求 index.html
- **WHEN** 客户端请求 `/index.html` 或 `/`
- **THEN** 响应包含 `Cache-Control: no-cache` 头

#### Scenario: 请求 assets 目录下的 JS/CSS 文件
- **WHEN** 客户端请求 `/assets/index-abc123.js`
- **THEN** 响应包含 `Cache-Control: public, max-age=31536000, immutable` 头

#### Scenario: 请求其他静态文件
- **WHEN** 客户端请求 `/favicon.ico`
- **THEN** 响应包含 `Cache-Control: public, max-age=3600` 头
