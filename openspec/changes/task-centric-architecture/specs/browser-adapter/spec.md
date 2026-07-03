## ADDED Requirements

### Requirement: UI Action Plan 映射（Phase 2）

Browser Adapter 将 Execution Plan 映射为 UI Action Plan，再生成 Playwright 测试。

- UI Action Plan 是 `adapters/browser/mapping/<capability>/v<version>.yaml`
- UI Action Plan 包含 UI Intent 步骤（"Open Contract page"、"Click Collect Rent"），不含 Playwright 代码
- UI Action Plan 是**可缓存的推导结果，不是权威来源**
- UI Action Plan 版本化，与 Capability 版本绑定
- Knowledge Layer 永远高于 UI（Capability 变化时 UI mapping 需重新推导）

#### Scenario: 从 Capability 推导 UI Action Plan
- **WHEN** Browser Adapter 加载 collect-rent 的 Capability
- **THEN** 生成 UI Action Plan，包含从系统首页到完成收款的完整 UI 操作路径

### Requirement: 语义回放（Semantic Replay，Phase 2）

- Browser Adapter 验证 observability（Payment 是否存在）而非 steps 是否一致
- Browser 执行路径可以不同，但最终业务结果必须与 Execution Plan 的 observability 匹配

#### Scenario: UI 变化但语义不变
- **WHEN** UI 按钮布局变化但功能不变
- **THEN** Browser Adapter 仍可通过观察 observability 确认执行成功
- **THEN** CI 标记 UI drift 但不阻断

### Requirement: UI Mapping 版本化

- `adapters/browser/mapping/collect-rent/v1.yaml`、`v2.yaml`
- Execution Plan 中记录 `ui_mapping_version`

#### Scenario: UI 重构后版本升级
- **WHEN** collect-rent 页面重构
- **THEN** 创建 v2.yaml，CI 通过 semantic replay 验证 v1 baseline 和 v2 的 observability 一致
