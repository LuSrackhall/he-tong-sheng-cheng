## Why

合同模板中的 `${signingDate}`（签订日期）是用户的常见需求。当前系统只有 `${today}` 字段，但动态生成模式下每次下载日期都会变，不能代表签订时间。数据库中 Contract 实体已有 `CreatedAt` 字段（创建时间），语义等同于签订日期，但未暴露给字段映射系统。

## What Changes

- 在后端 `buildReplaceValues` 函数中添加 `signingDate` → `contract.CreatedAt.Format("2006-01-02")`
- 在前端 `presetFieldGroups` 的"合同类"分组中添加 `signingDate`，标签为"签订日期"
- 在 `PreviewTemplate` handler 的字段列表和内置字段集合中注册 `signingDate`

## Capabilities

### New Capabilities

（无新增能力，属于现有能力的扩展）

### Modified Capabilities

- `template-field-management`: 在预置字段组中新增 `signingDate` 字段（"合同类"分组），在后端模板预览和替换值中注册该字段

## Impact

- **后端**：`internal/transport/handler/template.go` — `buildReplaceValues` 函数 + `PreviewTemplate` handler
- **前端**：`frontend/src/views/Settings.vue` — `presetFieldGroups` + `presetFieldLabels`
- **无数据库变更**：CreatedAt 由 GORM 自动管理，无需迁移
- **无 API 变更**：复用现有字段映射机制
