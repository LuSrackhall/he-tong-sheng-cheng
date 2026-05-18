## Context

当前模板管理功能存在四个缺陷：
1. 模板无法删除
2. 用户无法添加自定义映射字段，只能使用预设的20个字段
3. 字段映射交互冗余：添加到 JSON 后还需手动点击开关启用
4. 字段用途不透明：只显示 `${fieldName}`，不知道在 UI 中代表什么

技术栈：Go + Gin 后端，Vue 3 + TypeScript 前端，GORM 操作 SQLite/PostgreSQL。

## Goals / Non-Goals

**Goals:**
- 实现模板删除功能，删除前检查模板是否被合同引用
- 支持用户添加自定义字段（字段名 + 显示标签）
- 简化启用逻辑：点击添加字段时自动启用
- 显示字段标签，让用户理解字段含义

**Non-Goals:**
- 不实现软删除（被使用的模板直接拒绝删除）
- 不新增数据表存储字段元数据（复用 fieldMap JSON）
- 不重构整个模板管理为表格形式

## Decisions

### 1. 模板删除策略

**决策**：硬删除，删除前检查 `Contract.TemplateID` 是否引用该模板。

**理由**：
- 模板一旦被合同使用，删除会导致合同导出失败
- 软删除增加复杂度，且被使用的模板应当阻止删除而非隐藏
- 实现简单：`SELECT COUNT(*) FROM contracts WHERE template_id = ?`

**替代方案**：软删除 + 自动清空关联合同的 templateId → 被否决，因为会破坏历史合同的可追溯性

### 2. 自定义字段存储

**决策**：复用现有 `fieldMap` JSON 字段，格式为 `{"fieldName": "displayLabel"}`。

**理由**：
- 无需数据库迁移
- 与现有预设字段格式一致
- 前端可直接读写 JSON

**替代方案**：新建 `template_fields` 表 → 被否决，过度设计

### 3. 启用逻辑简化

**决策**：添加字段时自动加入 `activeSet`（标记为启用）。所有启用的字段在上传 Word 时均需校验其占位符是否存在。开关控制字段是否参与替换和校验。

**理由**：
- 用户期望：添加到映射的字段就应该被使用
- 用户明确要求：只要是启用的，都需要校验
- 仍保留开关是因为有些预设字段（如 `assetDescription`）可能 Word 中没有，可临时禁用避免校验失败

### 4. 显示标签

**决策**：从 `fieldMap` 中读取标签值，UI 展示 `${fieldName} (显示标签)`。

**理由**：
- 让用户理解 `tenantName` 对应"租户姓名"
- 预设字段的标签硬编码，自定义字段的标签由用户输入

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 被使用的模板删除会导致孤儿数据 | 删除前查询 `contracts` 表，存在引用则返回 409 |
| 自定义字段名与预设字段冲突 | 前端校验不允许重复字段名 |
| fieldMap JSON 格式错误导致解析失败 | 保存前校验 JSON 有效性 |

## Migration Plan

无需数据库迁移。

1. 后端：添加 DELETE 端点，在 TemplateRepo 中添加 `IsUsedByContract` 和 `Delete` 方法
2. 前端：
   - Settings.vue 添加删除按钮
   - 添加"添加自定义字段"弹窗
   - 修改 `insertFieldPlaceholder`：添加字段后自动调用 `toggleActive` 启用
   - 修改预设字段展示：从 fieldMap 读取标签

## Open Questions

- 是否支持编辑已有字段的标签？→ 用户可手动编辑 JSON 实现，暂不做 UI
