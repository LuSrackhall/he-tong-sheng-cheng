## Context

Settings.vue 的 field-chip 区域（约第 806-829 行）渲染所有预置和自定义字段标签。当前每个 chip 由两部分组成：chip-label（点击添加/移除映射）和 chip-toggle（启用/禁用开关）。需要在不改变现有交互的前提下增加复制功能。

前端技术栈：Vue 3 + TypeScript + Vite，无 UI 组件库，所有样式均为局部 scoped CSS。

## Goals / Non-Goals

**Goals:**
- 在 chip-label 和 chip-toggle 之间插入独立复制按钮
- 使用 `navigator.clipboard.writeText()` 复制 `${chip.name}` 到剪贴板
- 复制后显示 Tooltip "已复制"，1.5 秒后自动消失
- 事件隔离：`@click.stop` 阻止冒泡到 chip-label

**Non-Goals:**
- 不提取 composable（功能简单，无需复用）
- 不添加 clipboard polyfill
- 不修改映射/切换逻辑

## Decisions

### 1. DOM 结构

将复制按钮作为 `.field-chip` 的直接子元素，放在 `.chip-label` 和 `.chip-toggle` 之间。避免 button 嵌套问题。

```
<div class="field-chip">
  <button class="chip-label">...</button>      <!-- 现有 -->
  <button class="chip-copy">...</button>        <!-- 新增 -->
  <button class="chip-toggle">...</button>      <!-- 现有 -->
</div>
```

### 2. Tooltip 状态管理

使用 `ref<Set<string>>` 跟踪哪些 chip 刚刚被复制过（`recentlyCopied`）。复制时将 chip.name 加入 Set，1.5 秒后移除。模板中用 `v-if="recentlyCopied.has(chip.name)"` 控制 Tooltip 显示。

选择 Set 而非单个字符串，因为理论上用户可能快速连续点击多个 chip。

### 3. 复制图标

使用内联 SVG（clipboard 图标，约 3 行），尺寸 12x12px，与 chip 整体风格一致。

### 4. 错误处理

`navigator.clipboard.writeText()` 在非安全上下文下会 reject。catch 错误后静默处理（此系统为内部工具，部署环境可控），不弹出错误提示干扰用户。

## Risks / Trade-offs

- **[风险] navigator.clipboard 在 HTTP 环境不可用** → catch 静默降级，不影响主流程
- **[权衡] Tooltip 使用绝对定位** → 需要 `.field-chip` 设为 `position: relative`，这是安全的 CSS 变更
- **[权衡] 新增 `.field-chip` 子元素可能影响 flex 布局** → `.field-chip` 已有 `display: inline-flex`，新增子元素会自动参与 flex 排列
