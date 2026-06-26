# UX 审计修复设计文档

## Context

资产租赁与催缴管理系统前端（Vue 3 + TypeScript + Pinia）存在多处 UX 缺陷。后端 API 已完善，本次变更仅涉及前端。核心问题：

- 5 个主要列表页面（AssetList、TenantList、ContractList、ReceiptBookList、CollectRent）缺少加载/错误/空状态
- API 调用缺少错误处理，错误被静默吞掉
- `useDebounce` 在 4 个文件中完全重复定义
- Modal 缺少基础无障碍属性
- 部分状态标签显示英文而非中文
- 无动态页面标题
- 危险操作使用原生 `confirm()` 对话框

## Goals / Non-Goals

**Goals:**
- 所有列表页面具备完整的加载中/错误/空状态展示
- 增强 Axios 拦截器处理 5xx/网络错误，同时各 view 保留本地 try/catch
- 创建可复用的 ConfirmDialog 组件替代原生 `confirm()`
- 为所有 Modal 添加基础无障碍属性（role="dialog"、aria-modal、aria-labelledby）
- 修复状态标签英文显示问题
- 通过 router afterEach 为每个页面设置动态 document.title
- 提取 `useDebounce` 为共享 composable
- 为已有搜索输入框添加清除按钮

**Non-Goals:**
- 不添加后端 DELETE 端点
- 不拆分 Settings 页面
- 不添加面包屑导航
- 不添加深色模式
- 不添加 ReceiptList 搜索（需要后端支持，单独 change）
- 不实现完整 focus trap（scope 太大）

## Decisions

### 1. 分层错误处理

增强 Axios 响应拦截器处理 5xx 和网络错误（toast 提示），各 view 保留本地 try/catch 设置 `error` 状态变量用于内联错误 UI。两层职责不同：拦截器防止静默失败，view catch 提供上下文相关的用户体验。

### 2. 可复用 ConfirmDialog 组件

创建 `frontend/src/components/ConfirmDialog.vue`，props: title, message, confirmText, variant。使用 v-model:visible + emit 模式。遵循现有 modal overlay 样式。内部包含 role="dialog"、aria-modal="true"、aria-labelledby。替换 Settings.vue、UserManagement.vue、NewContract.vue、CollectRent.vue 中的 6 处 `confirm()` 调用。

### 3. Modal 无障碍

除 ConfirmDialog 内置外，为所有现有 Modal overlay 添加 `role="dialog"`、`aria-modal="true"`、`aria-labelledby` 属性。纯 additive HTML 属性变更，无回归风险。

### 4. document.title

在 router afterEach 守卫中设置 `document.title = to.meta.title + ' - 租赁管家'`。每个路由定义添加 `meta: { title: '...' }`。

### 5. 搜索清除按钮

为已有搜索输入框添加 X 清除图标（行内 SVG），点击清空搜索词并触发搜索。

## Task Groups

### Task Group 1: 基础设施（组件 + composable + 拦截器）
- 创建 ConfirmDialog.vue 组件
- 创建 useDebounce.ts composable
- 增强 Axios 拦截器错误处理
- 添加 router afterEach 页面标题

### Task Group 2: 列表页面加载/错误/空状态
- AssetList.vue
- TenantList.vue
- ContractList.vue
- ReceiptBookList.vue
- CollectRent.vue

### Task Group 3: Modal 无障碍 + ConfirmDialog 替换 + 状态标签修复
- 为所有 Modal 添加无障碍属性
- 替换 confirm() 为 ConfirmDialog
- 修复英文状态标签

### Task Group 4: 搜索优化 + 代码清理
- 搜索清除按钮
- 替换内联 useDebounce 为共享 composable
- 修复 v-if/v-else 模式（AssetList、TenantList）

## Risks / Trade-offs

- **error handling 变更中风险** — CollectRent（支付创建+作废流程）、ContractList（复杂详情弹窗）、NewContract（多步向导）是最脆弱的 view，需确保 toast 和内联错误不重复触发
- **ConfirmDialog 替换低风险** — 纯 UI 替换，逻辑不变
- **ReceiptList 搜索已推迟** — 服务端分页下客户端过滤仅限当前页，需要后端支持搜索参数

## Files to Modify

**新建：**
- `frontend/src/components/ConfirmDialog.vue`
- `frontend/src/composables/useDebounce.ts`

**修改：**
- `frontend/src/api/index.ts` — 拦截器增强
- `frontend/src/router/index.ts` — meta.title + afterEach
- `frontend/src/views/AssetList.vue` — 加载/错误/空状态 + 搜索清除 + useDebounce
- `frontend/src/views/TenantList.vue` — 加载/错误/空状态 + 搜索清除 + useDebounce
- `frontend/src/views/ContractList.vue` — 加载/错误/空状态 + 搜索清除 + useDebounce + 状态标签
- `frontend/src/views/ReceiptBookList.vue` — 加载/错误/空状态
- `frontend/src/views/CollectRent.vue` — 加载/错误/空状态 + useDebounce + ConfirmDialog
- `frontend/src/views/Settings.vue` — ConfirmDialog
- `frontend/src/views/UserManagement.vue` — ConfirmDialog
- `frontend/src/views/NewContract.vue` — ConfirmDialog + 状态标签
- `frontend/src/views/Home.vue` — Modal 无障碍
- `frontend/src/views/ArrearsList.vue` — Modal 无障碍
- `frontend/src/views/ReceiptList.vue` — Modal 无障碍
