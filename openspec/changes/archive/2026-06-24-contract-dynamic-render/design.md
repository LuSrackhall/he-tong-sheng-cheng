## Context

当前实现中，合同预览和下载分别走两条独立路径：
- 预览：`PreviewContract` → `pdf.GenerateContractHTML` → 硬编码 HTML 模板
- 下载：`ExportContract` 生成缓存文件 → `DownloadContract` 读取文件返回

核心占位符替换逻辑在 `internal/docx/render.go` 的 `Render` 函数中，已验证可用。
替换值由 `buildReplaceValues`（`template.go:493`）统一构建。

## Goals / Non-Goals

**Goals:**
- 预览和下载都基于上传的 Word 模板动态生成
- 新增 `docx.ToHTML` 函数将 Word 内容转换为 HTML
- 合并 export + download 为单步操作

**Non-Goals:**
- 不实现 Word 到 HTML 的完美排版还原
- 不修改 `docx.Render` 的占位符替换逻辑
- 不改变模板上传和验证流程

## Decisions

### 1. docx → HTML 转换的实现方式

**选择**：在 `internal/docx/tohtml.go` 中实现，直接解析 docx ZIP 中的 `word/document.xml`，提取段落和表格，转换为 HTML。

**解析范围**：
- `<w:p>` → `<p>`（段落）
- `<w:r>` + `<w:t>` → 内联文本
- `<w:rPr>` 中的 `<w:b/>` → `<strong>`，`<w:i/>` → `<em>`
- `<w:tbl>` → `<table>`，`<w:tr>` → `<tr>`，`<w:tc>` → `<td>`
- `<w:pPr>/<w:jc>` → text-align 对齐
- `<w:pPr>/<w:ind>` → text-indent 缩进

**不支持**：图片、浮动框、分栏、页眉页脚样式、复杂嵌套表格。

**理由**：docx XML 结构规范，手动解析可控且无需外部依赖。对于合同文档（以文字和简单表格为主）已足够。

**替代方案**：
- 引入 `unioffice` 库——依赖过重（支持完整 OOXML），且我们只需要读取子集
- 用正则解析 XML——不可靠，XML 有嵌套结构

### 2. 接口改造方案

**DownloadContract** 改造：
```
原流程：检查缓存文件 → 返回文件
新流程：读取合同 → 读取模板 → docx.Render → 返回字节流
```

**PreviewContract** 改造：
```
原流程：读取合同 → buildReplaceValues → pdf.GenerateContractHTML → 返回 HTML
新流程：读取合同 → 读取模板 → docx.Render → docx.ToHTML → 返回 HTML
```

**ExportContract**：删除该接口及对应前端调用。

### 3. PreviewContract 返回的 HTML 结构

返回的 HTML 需要包含：
- 基础 CSS 样式（字体、页面宽度、打印支持、表格边框）
- 提示条："此为预览，精确格式请下载 Word 文件查看"
- Word 内容转换的 HTML 主体

使用 `contractPreviewWrapper` 常量（`template.go:104`），通过 `fmt.Sprintf` 注入 body HTML。

### 4. 前置校验逻辑复用

ExportContract 中有完善的校验逻辑（模板存在性、必填字段检查等）。提取为 `validateContractForExport` 私有方法（`template.go:132`），PreviewContract 和 DownloadContract 共用。

### 5. 无模板时的处理

合同未关联模板或模板文件未上传时，返回 `400 Bad Request`，错误信息："该合同未关联模板，请先在设置中上传并关联模板"。

### 6. docx XML 命名空间处理策略

**背景**：docx XML 使用 `w:` 命名空间前缀，但 `xml.Unmarshal` 的 `innerxml` 提取不会保留祖先元素的 `xmlns` 声明。这导致从表格/段落提取的内容中 `w:` 前缀成为未声明前缀。

**选择**：使用 `xml.RawToken` 代替 `xml.Token` 解析，配合 `stripPrefix` 函数（`tohtml.go:208`）剥离前缀后匹配元素名。

**理由**：
- `RawToken` 返回原始 XML token，不做命名空间解析
- 避免因 `xmlns` 声明缺失导致 local name 带前缀（`w:r` vs `r`）的匹配问题
- `stripPrefix` 统一处理两种路径（原始 XML 已解析 vs 重建 XML 未解析）

### 7. 错误信息脱敏

**选择**：Handler 返回给客户端的错误信息不包含内部路径或底层错误详情，仅返回通用提示（如"生成合同文件失败，请重试"）。完整错误通过 `log.Printf` 记录在服务端日志中。

## Risks / Trade-offs

- **[HTML 还原度]** → 复杂 Word 排版无法在 HTML 中完美呈现。Mitigation：添加提示文案引导用户下载 Word 查看精确格式。
- **[XML 命名空间]** → docx XML 使用多个命名空间（w、r、wp 等），解析时需正确处理。Mitigation：使用标准 `encoding/xml` 并注册常见命名空间前缀。
- **[性能]** → 每次请求动态生成，增加 CPU 开销。Mitigation：合同文档通常较小（几十 KB），生成耗时在毫秒级。

## Migration Plan

无数据迁移。部署后新接口立即生效。前端需同步更新以移除 export 步骤。
