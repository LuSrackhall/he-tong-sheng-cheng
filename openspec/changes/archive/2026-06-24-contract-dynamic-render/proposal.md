## Why

合同预览使用的是硬编码 HTML 模板，与用户上传的 Word 模板完全无关；合同下载依赖预先缓存的文件，日期等数据可能与系统实际值不一致。需要让预览和下载都基于上传的 Word 模板动态生成，确保内容准确一致。

## What Changes

- **预览接口改造**（`GET /contracts/:id/preview`）：从硬编码 HTML 改为读取 Word 模板 → 替换占位符 → 转换为 HTML 返回
- **下载接口改造**（`GET /contracts/:id/download`）：从读取缓存文件改为动态生成 Word 文件直接返回
- **移除 export 接口**（`POST /contracts/:id/export`）：不再需要单独的导出步骤，download 即生即下
- **新增 docx → HTML 转换能力**（`internal/docx/tohtml.go`）：将 Word 文档内容转换为浏览器可展示的 HTML
- **无模板时的行为统一**：预览和下载在合同无模板时返回明确错误提示

## Capabilities

### New Capabilities
- `docx-to-html`: 将 .docx 文件内容转换为 HTML，支持段落、表格、粗体/斜体、对齐等基本排版

### Modified Capabilities

（无已有 spec 需要修改）

## Impact

- **后端 handler**：`internal/transport/handler/template.go` — 重写 `PreviewContract` 和 `DownloadContract`，删除 `ExportContract`
- **新增文件**：`internal/docx/tohtml.go` — docx 转 HTML 实现
- **已删除**：`internal/pdf/contract.go` 中的 `ContractData` 和 `contractHTML` 模板（不再使用，仅保留 TemplatePreview 相关代码）
- **前端 API**：`frontend/src/api/index.ts` — 移除 `contractApi.export()` 调用，下载流程简化
- **前端视图**：`NewContract.vue`、`ContractList.vue` — 移除 export 步骤，下载直接调 download 接口
- **API 路由**：`cmd/server/main.go` — 移除 `POST /contracts/:id/export` 路由
