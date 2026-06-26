## Why

合同创建表单（NewContract.vue Step 3）错误地使用 `activeFields`（Word 校验开关）来决定哪些字段渲染为 UI 表单组件。管理员关闭某个字段的 Word 校验后，该字段会从合同创建表单中消失，导致用户无法填写。

## What Changes

1. **修复 activeFieldsList 数据源**：从 `fieldMap`（字段映射 JSON）获取所有字段列表，而非仅 `activeFields` 中为 `true` 的字段
2. **修复标签优先级**：`getFieldLabel` 优先从 `fieldMap` 读取用户自定义的中文标签，硬编码标签作为 fallback
3. **新增 fieldMap 解析逻辑**：解析 `selectedTemplate.value.fieldMap` JSON，失败时 fallback 到 `activeFields` 的键列表（兼容旧数据）
4. **表单提交适配**：`createContract` payload 需要收集自定义字段的值

## Capabilities

### Modified Capabilities
- `template-driven-contract-form`: Step 3 表单字段数据源从 activeFields 改为 fieldMap；标签优先级从硬编码改为 fieldMap 优先

## Impact

- **前端**：`frontend/src/views/NewContract.vue` — activeFieldsList computed、getFieldLabel 函数、createContract payload
- **后端**：已确认 `contractApi.create` 使用固定结构体，不支持自定义字段。自定义字段值在 payload 中传递但被后端忽略
- **兼容性**：旧模板（无 fieldMap）通过 fallback 到 activeFields 键列表保持兼容
