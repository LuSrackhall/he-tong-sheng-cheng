## Why

当前模板管理存在四个用户体验缺陷：模板无法删除导致管理混乱；用户只能使用预设的 20 个字段，无法添加自定义映射字段；字段添加到 JSON 后还需手动点击开关启用，交互冗余；字段只显示 `${fieldName}` 占位符，用户不知道它在 UI 中代表什么。这四个问题直接影响模板管理的可用性，需要现在处理。

## What Changes

**模板删除**
- From: 无删除功能
- To: DELETE `/api/templates/:id`，删除前查询 `contracts` 表，若存在引用返回 409，否则物理删除
- Reason: 模板一旦被合同使用，删除会导致导出失败
- Impact: 后端新增删除端点和检查方法，前端新增删除按钮

**自定义字段映射**
- From: 仅预设 20 个字段，用户不可扩展
- To: 增加"添加自定义字段"按钮，弹出表单输入字段名和显示标签，新字段写入 `fieldMap` JSON 并自动追加到预设列表
- Reason: 不同场景需要不同字段，预设字段无法覆盖所有需求
- Impact: 前端新增弹窗组件，后端无需改动（复用现有 `fieldMap` 存储）

**启用逻辑简化**
- From: 点击字段标签添加到 `fieldMap` 后，还需手动点击开关添加到 `activeFields` 才启用
- To: 添加字段时自动加入 `activeSet`（标记为启用）；所有启用的字段在上传 Word 时均需校验其占位符是否存在；关闭开关则该字段不参与替换也不参与校验
- Reason: 用户期望添加到映射的字段就应该被使用和校验
- Impact: 前端修改 `insertFieldPlaceholder` 逻辑；上传校验逻辑改为对所有 active 字段进行校验

**字段标签展示**
- From: 只显示 `${fieldName}`
- To: 显示 `${fieldName} → 显示标签`
- Reason: 让用户理解 `tenantName` 对应"租户姓名"，字段用途透明化
- Impact: 前端预设字段标签改为从 `fieldMap` 读取标签值展示

## Capabilities

### New Capabilities
- `template-deletion`: 模板删除功能，含合同引用检查，防止误删已使用的模板
- `template-field-management`: 自定义字段映射（添加/标签）、自动启用、全量校验控制

### Modified Capabilities
<!-- 本次变更不修改已有 capability spec -->

## Impact

| 层级 | 影响 |
|------|------|
| 后端 handler | `template.go` 新增 Delete；`contract.go` 的上传校验改为对所有 active 字段校验 |
| 后端 domain/repo | `TemplateRepo` 接口新增 `Delete`、`IsUsedByContract` 方法 |
| 前端 Settings.vue | 新增删除按钮、"添加自定义字段"弹窗、字段标签展示、自动启用逻辑 |
| 数据库 | 无变更（复用 `fieldMap` JSON 列） |
| API | 新增 `DELETE /api/templates/:id` |
