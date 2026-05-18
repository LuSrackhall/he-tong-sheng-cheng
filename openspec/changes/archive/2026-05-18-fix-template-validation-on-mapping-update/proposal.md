# Proposal: fix-template-validation-on-mapping-update

> Written: 2026-05-18
> Based on: [brainstorm.md](brainstorm.md), [design.md](design.md)

---

## Why

上轮 change `improve-template-management` 引入了 docx placeholder 校验逻辑，但两个 bug 导致校验实际未生效：(1) `UploadTemplate` 用了 `io.Reader.Read` 而非 `io.ReadAll`，文件读取不完整；(2) `UpdateTemplateMapping` 修改映射后完全未触发重校验。此外，用户明确要求“启用的字段必须校验”且“更改映射后提示用户 Word 文件是否符合新要求”。本次修复补全校验链路并新增 `Validated` 状态字段，使得使用时（导出合同）可阻止未通过校验的模板。

## What Changes

**UploadTemplate 文件读取修复**
- From: `src.Read(fileData)` 只读一次缓冲区，大文件读取不完整
- To: `io.ReadAll(src)` 保证完整读取整个文件
- Reason: `io.Reader.Read` 不保证一次读完所有数据
- Impact: 非破坏性修复

**Template 新增 Validated 状态字段**
- From: Template 无校验状态
- To: 新增 `Validated bool` 字段，上传/修改映射时自动更新
- Reason: 需要在“保存时校验”和“使用时阻止”之间桥接
- Impact: 新增数据库列（GORM AutoMigrate），已有模板默认 `false`

**UpdateTemplateMapping 保存后重校验**
- From: 只保存 fieldMap/activeFields，不校验已上传文件
- To: 保存后若有已上传文件则重新校验并更新 Validated
- Reason: 映射变更后 Word 文件 placeholder 可能不再满足 activeFields
- Impact: 后端响应多返回 `validated` 字段

**ExportContract 校验门**
- From: 导出时不检查校验状态
- To: Validated=false 时返回 409 阻止导出，提示“请先上传符合要求的 Word 文件”
- Reason: 用户要求使用时阻止未通过校验的模板
- Impact: 破坏性——已有模板需重新上传 Word 触发校验后才能导出

## Capabilities

### New Capabilities
- `template-validation-status`: 模板校验状态管理——上传时校验并记录状态，修改映射时重新校验，导出时阻止未通过校验的模板

### Modified Capabilities
- `template-field-management`: 修改映射保存后新增重校验行为

## Impact

| 层级 | 影响范围 |
|------|---------|
| Domain | `internal/domain/template.go` — 新增 `Validated bool` 字段 |
| Handler | `internal/transport/handler/template.go` — 修复 `io.ReadAll` |
| Handler | `internal/transport/handler/contract.go` — `UpdateTemplateMapping` 加重校验, `ExportContract` 加 Validated 检查 |
| Frontend | `frontend/src/api/index.ts` — Template 类型新增 `validated` 字段 |
| Frontend | `frontend/src/views/Settings.vue` — 展示校验状态提示 |
| DB | `templates` 表新增 `validated` 列（AutoMigrate） |
