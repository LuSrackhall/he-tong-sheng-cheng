## ADDED Requirements

### Requirement: UX Audit Mode（Phase 3）

Exploration Adapter 以 Agent（LLM + 浏览器）驱动的方式执行 Capability。

- Agent 接收 Capability 的 `goal` + `scenarios` 作为 Prompt 输入
- Agent 自主决定操作步骤，记录每一步
- Agent 输出 UX Report（路径、困惑点、耗时、UX Score）
- Agent 在 CI 中作为**异步后处理**运行，不阻断 CI

#### Scenario: Agent 完成收租金操作
- **WHEN** Exploration Adapter 以 UX 模式执行 collect-rent
- **THEN** Agent 从首页导航到收款页面，完成填写，提交
- **THEN** 输出 UX Report（步骤数、耗时、困惑点）

### Requirement: Persona 系统

- Exploration Adapter 支持 4 种 Persona：
  - 村委会工作人员（基础操作、无技术背景）
  - 财务人员（高频重复操作）
  - 管理员（权限管理、备份恢复）
  - 新员工（第一次使用、必须靠界面文字导航）

#### Scenario: 新员工 Persona 发现导航问题
- **WHEN** Exploration Adapter 以新员工 Persona 执行 sign-contract
- **THEN** Agent 记录无法立即找到"新建合同"入口的困惑点

### Requirement: Chaos Mode

- Chaos Mode 测试异常操作场景：重复提交、无效输入、F5 刷新、缺少前置条件等
- Chaos Mode 的结果标记为行为观察而非失败

#### Scenario: 重复提交测试
- **WHEN** Chaos Mode Agent 重复点击收款提交按钮
- **THEN** 系统不应创建重复 Payment
- **THEN** Agent 记录"系统正确处理了重复提交"

### Requirement: 探索预算

- Exploration Adapter MUST 强制执行 `exploration.budget`
- `budget = { max_steps: 200, max_time_ms: 300000, max_capabilities: 10 }`
- 超过 budget 时当前探索中止，已收集的数据有效

#### Scenario: Agent 超时中止
- **WHEN** Agent 探索超过 max_time_ms
- **THEN** 当前探索中止，已有数据写入 Report
