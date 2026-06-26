## ADDED Requirements

### Requirement: Dashboard 统计 API 端点

系统 SHALL 提供 `GET /api/dashboard/stats` 端点，返回当前业务关键指标。此端点 SHALL 需要有效的 JWT 认证。

#### Scenario: 获取仪表盘统计数据
- **WHEN** 已认证用户请求 `GET /api/dashboard/stats`
- **THEN** 返回 HTTP 200，响应体包含以下字段：
  - `activeContracts`（活跃合同数，status 为 active 或 arrears 的合同数量）
  - `monthlyRevenue`（本月收款金额，当月所有 payment 的 amount 之和）
  - `overdueContracts`（逾期合同数，status 为 arrears 的合同数量）
  - `newContractsThisMonth`（本月新增合同数，当月创建的合同数量）

#### Scenario: 无认证时拒绝访问
- **WHEN** 未携带有效 JWT 的请求访问 `GET /api/dashboard/stats`
- **THEN** 返回 HTTP 401

### Requirement: Dashboard 前端概览页

前端 SHALL 在默认路由 `/` 展示 Dashboard 概览页，包含四个统计卡片。

#### Scenario: 首页展示统计数据
- **WHEN** 已登录用户访问首页
- **THEN** 展示四个卡片分别显示：活跃合同数、本月收款金额、逾期合同数、本月新增合同数
- **THEN** 数据从 `GET /api/dashboard/stats` 加载

#### Scenario: 数据加载中
- **WHEN** 首页正在加载统计数据
- **THEN** 展示加载状态指示

#### Scenario: 数据加载失败
- **WHEN** Dashboard API 请求失败
- **THEN** 展示错误提示，不阻断其他页面功能

### Requirement: 默认路由指向 Dashboard

路由 `/` SHALL 指向 Dashboard 概览页而非签新合同页。

#### Scenario: 访问根路径
- **WHEN** 已登录用户访问 `/#/`
- **THEN** 展示 Dashboard 概览页
