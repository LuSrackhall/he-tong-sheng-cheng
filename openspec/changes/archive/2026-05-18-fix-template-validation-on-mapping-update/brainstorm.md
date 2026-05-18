# Brainstorm: fix-template-validation-on-mapping-update

> Written: 2026-05-18
> Status: approved

---

## Design Summary

修复两个校验问题：
1. `UploadTemplate` 中 `io.Reader.Read` 不保证一次读完，导致大文件校验不生效
2. `UpdateTemplateMapping` 修改映射后不重新校验已上传的 Word 文件

并在 Template 新增 `Validated` 状态字段，使得导出合同时可阻止使用未通过校验的模板。

---

## Alternatives Considered

### 方案 A：Template 新增 `Validated` 字段（推荐）
- **做法**：
  - Template struct 加 `Validated bool` 字段
  - `UploadTemplate`：修复 `io.ReadAll`，校验结果写入 `Validated`
  - `UpdateTemplateMapping`：有已上传文件时重新校验并更新 `Validated`
  - `ExportContract`：检查 `Validated`，false 则返回 409 阻止导出
- **优点**：校验状态显式可查，导出时 O(1) 判断，不需要每次解压 docx
- **缺点**：新增数据库字段，需要 migration
- **为何采用**：状态显式、性能好、语义清晰

### 方案 B：导出时实时校验，不存状态
- **做法**：不在数据库存校验状态，ExportContract 每次都重新校验
- **优点**：不改 schema，无状态同步问题
- **缺点**：每次导出都要解压 zip 读 XML，性能开销大；用户到导出时才发现问题，体验差
- **为何未采用**：性能差、用户体验差

---

## Agreed Approach

采用方案 A：
- 修复 `io.ReadAll` 保证校验生效
- 新增 `Validated` 字段存储校验状态
- `UpdateTemplateMapping` 保存后自动重校验
- `ExportContract` 阻止未通过校验的模板导出

---

## Key Decisions

1. **映射保存不阻止** — 保存 fieldMap/activeFields 总是成功，校验在后台进行并更新 Validated 状态
2. **使用时阻止** — 仅在导出合同（ExportContract）时检查 Validated，未通过则返回 409 并提示"模板映射校验未通过，请先上传符合要求的 Word 文件"
3. **校验时机** — 上传 Word 时校验 + 修改映射时（如有已上传文件）重校验
4. **Validated 默认值** — 新模板无文件时 `Validated=false`，上传通过后 `true`

---

## Open Questions

（无）
