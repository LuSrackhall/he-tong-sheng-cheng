## 1. docx → HTML 转换模块

- [x] 1.1 新建 `internal/docx/tohtml.go`，实现 `ToHTML(docxData []byte) (string, error)` 函数：解压 docx ZIP，解析 `word/document.xml`，将 `<w:p>` 段落转换为 `<p>`，`<w:r>/<w:t>` 文本保留为内联内容
- [x] 1.2 支持基本格式：`<w:b/>` → `<strong>`，`<w:i/>` → `<em>`，`<w:jc>` → text-align，`<w:ind firstLine>` → text-indent
- [x] 1.3 支持表格：`<w:tbl>` → `<table>`，`<w:tr>` → `<tr>`，`<w:tc>` → `<td>`
- [x] 1.4 新建 `internal/docx/tohtml_test.go`，覆盖：纯文本文档、带格式文本、表格、空文档、无效文件等场景
- [x] 1.5 运行 `go test ./internal/docx/... -count=1` 确保通过

## 2. 后端接口改造

- [x] 2.1 从 `ExportContract` 中提取前置校验逻辑为私有方法 `validateContractForExport`，供 Preview 和 Download 共用
- [x] 2.2 重写 `DownloadContract`：移除缓存文件读取，改为调用 `validateContractForExport` → 读取模板 → `docx.Render` → 直接返回字节流（Content-Type: application/vnd.openxmlformats-officedocument.wordprocessingml.document）
- [x] 2.3 重写 `PreviewContract`：移除硬编码 HTML 模板调用，改为 `validateContractForExport` → 读取模板 → `docx.Render` → `docx.ToHTML` → 包裹基础 CSS 后返回 HTML
- [x] 2.4 删除 `ExportContract` handler 方法
- [x] 2.5 从 `cmd/server/main.go` 中移除 `POST /contracts/:id/export` 路由注册
- [x] 2.6 清理不再使用的代码：`internal/pdf/contract.go` 中的 `ContractData` 和 `contractHTML` 常量（确认无其他引用后删除）
- [x] 2.7 运行 `go test ./... -count=1` 确保全部通过

## 3. 前端改造

- [x] 3.1 `frontend/src/api/index.ts`：移除 `contractApi.export()` 方法，简化下载流程为直接调用 download 接口（blob 下载）
- [x] 3.2 `frontend/src/views/NewContract.vue`：Step 4 成功后移除 export 步骤，"生成并下载合同"按钮直接调 download 接口
- [x] 3.3 `frontend/src/views/ContractList.vue`：详情弹窗中的"下载合同"按钮直接调 download 接口，移除任何 export 相关逻辑
- [x] 3.4 运行 `vue-tsc --noEmit` 和 `npm run build` 确保类型检查和构建通过

## 4. 集成验证

- [x] 4.1 运行 `go build ./...` 确保编译通过
- [ ] 4.2 启动服务，上传 Word 模板，创建合同，验证预览显示模板内容（非旧 HTML）
- [ ] 4.3 验证下载的 Word 文件包含正确的合同数据（日期、租户名等与系统一致）
- [ ] 4.4 验证无模板的合同请求预览/下载时返回明确错误提示

---

## Post-Implementation Workflow

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run `/opsx:verify` to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, go to main branch and merge (must ask user)
4. **Archive**: Run `/opsx:archive` on main
5. **Cleanup**: `git worktree remove .worktrees/change/<name>`
