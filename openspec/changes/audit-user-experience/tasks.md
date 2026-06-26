## 1. 基础设施 — 组件、Composable、拦截器、路由

- [ ] 1.1 创建 ConfirmDialog.vue 组件（props: title, message, confirmText, variant；v-model:visible；内置 role="dialog"、aria-modal、aria-labelledby）
- [ ] 1.2 创建 useDebounce.ts composable（从现有内联实现提取）
- [ ] 1.3 增强 Axios 响应拦截器：处理 5xx → toast"服务器错误，请稍后重试"，网络错误 → toast"网络连接失败，请检查网络"，错误继续 reject
- [ ] 1.4 路由添加 meta.title 和 afterEach 守卫设置 document.title

## 2. 列表页面加载/错误/空状态

- [ ] 2.1 AssetList.vue — 添加 loading/error/empty 三态 + 搜索清除按钮
- [ ] 2.2 TenantList.vue — 添加 loading/error/empty 三态 + 搜索清除按钮
- [ ] 2.3 ContractList.vue — 添加 loading/error/empty 三态 + 搜索清除按钮
- [ ] 2.4 ReceiptBookList.vue — 添加 loading/error/empty 三态
- [ ] 2.5 CollectRent.vue — 添加 loading/error/empty 三态

## 3. Modal 无障碍 + ConfirmDialog 替换 + 状态标签

- [ ] 3.1 为所有现有 Modal 添加 role="dialog"、aria-modal="true"、aria-labelledby 属性
- [ ] 3.2 替换 CollectRent.vue 的 confirm() 为 ConfirmDialog（作废付款）
- [ ] 3.3 替换 UserManagement.vue 的 confirm() 为 ConfirmDialog（删除用户）
- [ ] 3.4 替换 Settings.vue 的 confirm() 为 ConfirmDialog（删除模板、恢复备份）
- [ ] 3.5 替换 NewContract.vue 的 confirm() 为 ConfirmDialog
- [ ] 3.6 修复 ContractList.vue 和 NewContract.vue 中英文状态标签为中文

## 4. 代码清理 — 替换内联 useDebounce

- [ ] 4.1 AssetList.vue — 替换内联 useDebounce 为共享 composable
- [ ] 4.2 TenantList.vue — 替换内联 useDebounce 为共享 composable
- [ ] 4.3 ContractList.vue — 替换内联 useDebounce 为共享 composable
- [ ] 4.4 CollectRent.vue — 替换内联 useDebounce 为共享 composable
- [ ] 4.5 AssetList.vue/TenantList.vue — 修复 v-if/v-else 模式

## 5. 构建验证

- [ ] 5.1 运行 vue-tsc --noEmit 类型检查通过
- [ ] 5.2 运行 npm run build 前端构建通过

---

## Post-Implementation Workflow

<!-- DO NOT MODIFY THIS SECTION -->

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run `/opsx:verify` to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, go to main branch and merge (must ask user)
4. **Archive**: Run `/opsx:archive` on main
5. **Cleanup**: `git worktree remove .worktrees/change/<name>`
