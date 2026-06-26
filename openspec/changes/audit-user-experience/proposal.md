## Why

前端多个核心列表页面（资产、租户、合同、收据本、收款）缺少加载/错误/空状态处理，API 调用错误被静默吞掉，用户在操作失败时得不到任何反馈。此外，Modal 缺少无障碍属性、危险操作使用原生 `confirm()` 对话框、部分状态标签显示英文、页面标题不更新等问题影响整体用户体验。

## What Changes

- 增强 Axios 响应拦截器，处理 5xx 和网络错误（toast 提示），各 view 保留本地 try/catch
- 为 5 个列表页面添加加载中/错误/空状态三态展示
- 创建可复用 ConfirmDialog.vue 组件，替换 6 处原生 `confirm()` 调用
- 为所有 Modal 添加基础无障碍属性（role="dialog"、aria-modal、aria-labelledby）
- 修复合同状态标签英文显示问题
- 通过 router afterEach 为每个页面设置动态 document.title
- 提取 useDebounce 为共享 composable，消除 4 处重复代码
- 为已有搜索输入框添加清除按钮

## Capabilities

### New Capabilities
- `loading-error-empty-states`: 列表页面加载中/错误/空状态三态展示机制
- `confirm-dialog`: 可复用确认对话框组件，替代原生 confirm()
- `modal-accessibility`: Modal 基础无障碍属性（role、aria-modal、aria-labelledby）
- `error-handling-interceptor`: Axios 拦截器增强，处理 5xx 和网络错误

### Modified Capabilities
<!-- 无现有 spec 需要修改 -->

## Impact

- **前端文件**: 13 个 Vue view 文件修改，1 个新组件，1 个新 composable，1 个 API 拦截器修改，1 个 router 修改
- **风险**: CollectRent（支付流程）、ContractList（复杂弹窗）、NewContract（多步向导）是最脆弱的 view，需确保 toast 和内联错误不重复触发
- **无后端变更**: 所有修改仅涉及前端
