## Context

前端 12 个 Vue view 中，5 个列表页面缺少加载/错误/空状态处理，API 错误被 Axios 拦截器静默吞掉（仅处理 401）。`useDebounce` 在 4 个文件中重复定义。6 处使用原生 `confirm()` 进行危险操作确认。所有 Modal 缺少无障碍属性。Home.vue 和 ArrearsList.vue 已有加载/错误状态模式，可作为参考实现。

## Goals / Non-Goals

**Goals:**
- 所有列表页面具备加载中/错误/空状态三态展示
- 分层错误处理：Axios 拦截器 + view 级 try/catch
- 可复用 ConfirmDialog 组件替代原生 confirm()
- Modal 基础无障碍
- 动态页面标题
- 消除 useDebounce 重复代码

**Non-Goals:**
- 不修改后端 API
- 不添加 focus trap
- 不拆分 Settings 页面
- 不添加深色模式

## Decisions

### 1. 错误处理分层

**选择**: 增强 Axios 响应拦截器 + 各 view 保留本地 try/catch

**理由**: 拦截器负责全局兜底（5xx、网络错误 → toast），view catch 负责上下文 UI（设置 error 状态变量 → 内联错误提示）。两层职责互补，不重复。

**替代方案**: 仅用拦截器（view 无法展示上下文相关错误 UI）、仅用 try/catch（需每个 view 重复错误处理逻辑）

### 2. ConfirmDialog 组件设计

**选择**: 单个可复用 `ConfirmDialog.vue`，props: title, message, confirmText, variant('danger'|'default')

**模式**: v-model:visible + @confirm/@cancel emit。遵循现有 modal overlay 样式。内置无障碍属性。

**替代方案**: 各 view 内联确认弹窗（代码重复 6 次）

### 3. 加载/错误/空状态实现方式

**选择**: 各 view 内联条件渲染（与 Home.vue、ArrearsList.vue 保持一致），不创建独立组件

**理由**: 项目体量不需要通用组件，内联实现与现有模式一致，维护成本低

### 4. 页面标题

**选择**: router afterEach 守卫 + meta.title

**替代方案**: composable + watch（需要每个页面挂载逻辑，过度工程化）

### 5. useDebounce 提取

**选择**: 提取到 `frontend/src/composables/useDebounce.ts`，与 `useEscapeKey.ts` 并列

## Risks / Trade-offs

- [CollectRent 错误处理] → toast 和内联错误可能重复触发 → 拦截器使用 error.response 判断，toast 仅在无本地 catch 时触发
- [ConfirmDialog 替换] → 纯 UI 替换，逻辑不变，低风险
- [状态标签修复] → 需排查所有显示英文 status 的位置，可能遗漏边缘情况
