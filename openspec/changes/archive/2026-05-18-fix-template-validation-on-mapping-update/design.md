# Design: fix-template-validation-on-mapping-update

> Written: 2026-05-18
> Based on: [brainstorm.md](brainstorm.md)

---

## Context

上一轮 change `improve-template-management` 引入了 `ActiveFields` 和 docx placeholder 校验逻辑，但存在两个 bug：

1. **`UploadTemplate` 校验不生效** — `template.go:53` 使用 `src.Read(fileData)` 读取上传文件，Go 的 `io.Reader.Read` 不保证一次读完所有数据（只返回当前缓冲区可读的内容），docx 文件通常几十 KB 以上，大概率只读了部分内容，导致 `ValidatePlaceholders` 在 XML 中找不到 placeholder 而误判为"缺少字段"或者反之校验结果不可靠
2. **`UpdateTemplateMapping` 无校验** — 用户修改 fieldMap/activeFields 后，后端只保存数据到数据库，不检查已上传的 Word 文件是否包含新的 placeholder

此外，用户要求：
- 保存映射时不阻止（提示即可），但使用时（导出合同）要阻止未通过校验的模板

---

## Goals / Non-Goals

**Goals:**
1. 修复 `UploadTemplate` 的文件读取方式，用 `io.ReadAll` 确保完整读取
2. `UpdateTemplateMapping` 保存后自动对已上传文件重新校验
3. Template 新增 `Validated bool` 字段存储校验状态
4. `ExportContract` 在导出前检查 `Validated`，未通过则返回 409 阻止导出
5. 前端在导出失败时提示用户"模板校验未通过，请先上传符合要求的 Word 文件"

**Non-Goals:**
- 不阻止 fieldMap/activeFields 的保存操作
- 不新建 API 端点（复用现有端点，在内部加校验逻辑）
- 不处理空文件路径的模板（没有上传文件就不校验，Validated 保持 false）

---

## Decisions

### D1: 用 `io.ReadAll` 替代 `Read`

`io.ReadAll` 保证读完整个 reader，而 `Read` 只读一次缓冲区。这是校验不生效的根因。

### D2: Template 新增 `Validated` 字段

```go
type Template struct {
    // ... existing fields
    Validated bool `json:"validated" gorm:"default:false"`
}
```

默认 `false`，仅在以下场景更新：
- 上传文件且校验通过 → `true`
- 上传文件且校验不通过 → `false`
- 修改映射且有已上传文件 → 重校验后更新

### D3: UpdateTemplateMapping 保存后重校验

调用链：保存 fieldMap/activeFields → 检查 `tpl.FilePath` 是否有值 → 有则读取文件并校验 → 更新 `Validated`

如果文件不存在（路径有值但文件被删除）→ 当作无文件处理，Validated 保持 false。

### D4: ExportContract 校验门

在 `ExportContract` 开头加入检查：
```go
if tpl.FilePath != "" && !tpl.Validated {
    c.JSON(409, gin.H{"error": "模板校验未通过，请先上传符合要求的 Word 文件"})
    return
}
```

`FilePath == ""` 时不阻止（没有文件=无需校验），但这种情况后续会因"Template file not uploaded yet"报错，行为不变。

### D5: 前端处理

- 上传 Word 时：后端返回的 `Template` 已包含 `validated` 字段，前端展示校验结果
- 保存映射时：后端返回的 `Template` 包含更新后的 `validated`，前端展示提示
- 导出合同时：409 → 弹出提示"模板校验未通过，请先上传符合要求的 Word 文件"

---

## Risks / Trade-offs

- **[Risk] 新增数据库字段需要 migration** → GORM AutoMigrate 会自动加列，SQLite/PostgreSQL 都兼容，零停机
- **[Risk] 已有模板 Validated 默认为 false** → 已有模板如果没有重新上传过 Word 文件，导出时会被阻止。这是预期行为：用户需要重新上传 Word 文件触发校验才能继续使用
- **[Trade-off] 校验状态是快照** → 文件可能在校验后被替换（直接操作文件系统），但本系统没有文件外部修改的场景，可以接受

---

## Migration Plan

1. 修改 `Template` struct 添加 `Validated bool` 字段
2. GORM AutoMigrate 在服务启动时自动加列
3. 已有模板 `Validated` 默认为 `false`，用户需要重新上传 Word 文件触发校验

---

## Open Questions

（无）
