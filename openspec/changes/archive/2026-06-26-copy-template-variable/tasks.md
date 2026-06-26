## 1. Script 逻辑：复制函数与 Tooltip 状态

- [x] 1.1 在 `<script setup>` 中添加 `recentlyCopied = ref<Set<string>>(new Set())` 状态管理
- [x] 1.2 添加 `copyToClipboard(templateId: number, chipName: string)` 函数：调用 `navigator.clipboard.writeText('${' + chipName + '}')`，成功后将 chipName 加入 recentlyCopied Set，1.5 秒后移除；catch 错误静默处理

## 2. 模板：添加复制按钮元素

- [x] 2.1 在 `.field-chip` 容器内、`.chip-label` 和 `.chip-toggle` 之间插入复制按钮 `<button class="chip-copy">`，包含内联 SVG clipboard 图标
- [x] 2.2 绑定 `@click.stop="copyToClipboard(t.id, chip.name)"` 事件，`title="复制占位符"` 提示文本
- [x] 2.3 添加 Tooltip 元素 `<span class="copy-tooltip" v-if="recentlyCopied.has(chip.name)">已复制</span>`，作为 `.field-chip` 的子元素

## 3. 样式：复制按钮与 Tooltip CSS

- [x] 3.1 添加 `.chip-copy` 样式：12x12px 图标按钮，无边框无背景，与 chip 整体风格一致，hover 时显示浅色背景
- [x] 3.2 添加 `.copy-tooltip` 样式：绝对定位（相对于 .field-chip），深色背景白色文字，圆角，fade in/out 过渡动画
- [x] 3.3 确保 `.field-chip` 已有 `position: relative`（用于 Tooltip 定位）

---

## Post-Implementation Workflow

After completing ALL tasks above, follow this sequence strictly:

1. **Verify**: Run myspec-verify to produce verify.md
2. **User Acceptance**: Present change summary, ask user to confirm the problem is solved
3. **Merge**: After user accepts, notify coordinator and wait for merge signal
4. **Archive**: Run myspec-merge after receiving merge signal
5. **Cleanup**: git worktree remove
