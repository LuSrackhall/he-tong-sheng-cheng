## Context

当前合同的预览和下载是两条完全独立的路径：

- **预览**（`GET /contracts/:id/preview`）：使用硬编码在 Go 代码中的 HTML 模板（`internal/pdf/contract.go`），与用户上传的 Word 模板无关
- **下载**（`GET /contracts/:id/download`）：读取预先生成并缓存的文件（`uploads/exports/contract_{id}.docx`），必须先调用 export 接口生成

这导致两个问题：
1. 预览内容与用户上传的 Word 模板格式完全不同
2. 下载的合同日期可能与系统数据不一致（缓存文件未更新，或 Word 占位符替换失败）

合同导出的核心逻辑在 `internal/docx/render.go` 中，通过解压 docx（ZIP 格式）、遍历 XML 文件、替换 `${key}` 占位符实现。`buildReplaceValues` 函数（`template.go:493`）负责从合同/租户/资产数据构建替换值。

## Goals / Non-Goals

**Goals:**
- 预览和下载都基于用户上传的 Word 模板动态生成，确保数据与系统一致
- 每次请求都从数据库读取最新合同数据，实时替换占位符
- 预览在浏览器中展示 Word 模板的实际内容（转换为 HTML）
- 下载直接返回生成的 .docx 文件，无需预先 export

**Non-Goals:**
- 不改变 Word 模板的上传和验证逻辑
- 不改变占位符替换的核心算法（`docx.Render`）
- 不支持无模板的合同（无模板时返回明确错误提示）
- 不做 Word 到 HTML 的完美保真转换（保留基本排版：段落、表格、粗体/斜体、对齐即可）

## Decisions

### 1. 合并 export + download 为单一动态操作

**选择**：移除缓存文件机制，download 接口直接动态生成 Word 文件返回。

**理由**：
- 消除缓存过期/关联错误的问题
- 简化流程（用户不再需要先 export 再 download）
- 每次下载都使用最新数据

**替代方案**：保留 export + download 两步但强制每次 export 覆盖——不如直接合并简洁。

### 2. 预览采用 docx → HTML 转换

**选择**：新增 `internal/docx/tohtml.go`，从 docx XML 中提取内容并转换为 HTML，在浏览器中展示。

**理由**：
- 预览和下载使用同一个模板源，保证一致性
- 实现复杂度可控：docx XML 结构规范，提取段落/表格/基本格式即可
- 不引入外部重依赖（如 LibreOffice）

**替代方案**：
- 用第三方库（如 unioffice）——依赖过重，且我们只需要读取不需要编辑
- 保留当前 HTML 模板但同步内容——无法保证与 Word 模板一致

### 3. 无模板时的行为

**选择**：预览和下载接口在合同未关联模板或模板未上传时，返回 400 错误并提示用户先上传模板。

**理由**：强制使用模板确保一致性，避免维护两套生成逻辑。

### 4. PreviewContract 改造策略

**选择**：PreviewContract 改为先动态生成 Word，再调用 docx.ToHTML 转换为 HTML 返回。

**流程**：
1. 读取合同 + 模板（复用 ExportContract 的前置校验逻辑）
2. `docx.Render(templateData, values)` 生成 Word 数据（内存中，不写文件）
3. `docx.ToHTML(wordData)` 转换为 HTML
4. 返回 HTML 给前端

## Risks / Trade-offs

- **[docx → HTML 转换精度]** → Word 的排版特性远比 HTML 丰富，复杂格式（多列、浮动图片、页眉页脚样式）无法完美还原。Mitigation：接受基本排版（段落/表格/粗斜体/对齐/缩进），在页面提示"精确格式请下载 Word 文件查看"。
- **[性能]** → 每次请求都动态生成，比读缓存文件慢。Mitigation：docx 替换是纯内存操作，单次生成耗时约几十毫秒，对用户无感知。
- **[Word XML 命名空间]** → docx XML 使用 `w:` 命名空间前缀，但 `xml.Unmarshal` 的 `innerxml` 提取不保留 `xmlns` 声明。实际实现采用 `xml.RawToken` + `stripPrefix` 前缀剥离策略解决。
- **[错误信息安全]** → Handler 返回客户端的错误信息需脱敏，不暴露内部路径。完整错误记录在服务端日志。
