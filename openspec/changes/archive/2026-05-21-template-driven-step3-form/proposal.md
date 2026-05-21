## Why

当前签合同 Step 3 表单字段全部写死，与用户在 Settings 配置的模板字段映射完全脱节。用户花时间配置了 Word 模板的字段启用/禁用，但签合同时仍看到固定字段，无法根据模板灵活裁剪表单内容。同时缺少按字段独立控制 Word 校验的能力，当用户启用月租金+年租金两个字段但 Word 中只落一个时无法跳过校验。

## What Changes

- **BREAKING**: `activeFields` 从 `string[]` 升级为 `Record<string, boolean>`，`false` 表示启用但跳过 Word 校验
- Step 3 合同表单完全由模板 `activeFields` 决定展示哪些字段，不再写死
- 表单字段按值来源分类渲染：用户填写型（可编辑）、系统自动型（只读）、资产/租户选择型（只读）
- 新增 `yearlyRent` 预设字段，与 `monthlyRent` 在表单中支持联动换算，关联默认开启
- "已启用字段"摘要区 chip 增加校验独立开关，点击切换 `✓校验` / `✗不校验`，与 JSON 编辑器双向同步
- 向后兼容：旧 `activeFields` 的 `[]string` 格式自动转为 `Record<string, true>`

## Capabilities

### New Capabilities

- `template-driven-contract-form`: Step 3 合同表单根据模板 activeFields 动态渲染，字段按来源分类（可编辑/只读），月/年租金联动换算
- `per-field-validation-toggle`: 每个已启用字段可独立控制是否参与 Word 占位符校验，通过 chip 开关和 JSON 编辑器双重控制
- `yearly-rent-field`: 年租金作为独立预设字段加入模板映射，与月租金在表单中关联自动换算

### Modified Capabilities

- `template-field-management`: activeFields 从 `string[]` 升级为 `Record<string, boolean>`，保存映射时自动同步 activeFields
- `template-validation-status`: Word 校验规则从"所有已启用字段"变为"仅 activeFields[key] === true 的字段"

## Impact

- **前端**: `Settings.vue`（activeFields 处理 + 校验开关 chip UI）、`NewContract.vue`（Step 3 表单模板驱动渲染 + 月/年租金联动）
- **后端**: `internal/transport/handler/contract.go`（Create/UpdateTemplateMapping requiredFields 逻辑调整）、`internal/transport/handler/template.go`（parseActiveFields 兼容新旧格式 + 导出校验逻辑）、`internal/docx/validate.go`（校验逻辑适配新 activeFields）
- **数据迁移**: 现有模板的 `activeFields` 数组自动转为对象格式，无需手动操作
