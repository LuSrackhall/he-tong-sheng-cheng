## Context

用户在创建合同模板时，需要将变量占位符（如 `${tenantName}`）手动输入到 Word 文档中。当前 Settings.vue 页面的 field-chip 标签没有复制功能，用户只能逐字母手打，体验很差。

现有 chip 结构：chip-label（点击=添加/移除映射）+ chip-toggle（点击=启用/禁用）。需要在不改变现有交互的前提下增加复制功能。

## Goals / Non-Goals

**Goals:**
- 在每个 field-chip 标签上添加独立的复制按钮（小图标）
- 点击复制按钮将 `${chip.name}` 复制到剪贴板
- 复制后显示 Tooltip "已复制"，1-2 秒后自动消失
- 保持现有"点击标签文本=添加/移除映射"功能不变
- 零新依赖，纯 CSS + JS 实现
- 同时适用于预置字段和自定义字段

**Non-Goals:**
- 不修改映射配置逻辑
- 不添加批量复制功能
- 不引入 clipboard-polyfill 等第三方依赖
- 不修改 chip-toggle 的行为

## Decisions

### 1. 交互方式：独立复制按钮

在 chip-label 旁添加一个独立的复制图标按钮（SVG 内联图标），与现有点击区域物理分离。

**布局顺序**：`[chip-label 文本] [复制图标] [chip-toggle]`

复制按钮放在 field-chip 容器内，chip-label 和 chip-toggle 之间，使用 `@click.stop` 阻止事件冒泡。

### 2. 反馈形式：Tooltip

- 点击复制后，在复制按钮旁显示 `<span class="copy-tooltip">已复制</span>`
- 使用 CSS @keyframes 动画实现淡入淡出
- 1.5 秒后自动隐藏
- 使用 `navigator.clipboard.writeText()` API

### 3. 技术实现

- 复制逻辑直接写在 Settings.vue 的 script setup 中（功能简单，不需要提取 composable）
- Tooltip 状态使用 ref 按 chip.name 管理（已复制的 chip 名集合）
- 图标使用内联 SVG（clipboard 图标），约 3 行 SVG 代码
- 使用 `@click.stop` 阻止事件冒泡到 chip-label 的映射操作

## Risks / Trade-offs

- **[风险] navigator.clipboard 在非 HTTPS 环境不可用** → 此系统为内部管理工具，部署环境可控。如果后续遇到问题，可降级为 `document.execCommand('copy')` 兼容方案
- **[权衡] 复制按钮位置** → 放在 chip-label 和 chip-toggle 之间，保持 chip 整体视觉平衡
- **[风险] Tooltip 在小屏幕上可能被截断** → 使用 `position: absolute` 相对于复制按钮定位，自动适配
