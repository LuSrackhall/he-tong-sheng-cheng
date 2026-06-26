## Why

用户在创建合同模板时需要将变量占位符（如 `${tenantName}`）手动输入到 Word 文档中。当前 Settings.vue 的 field-chip 标签没有复制功能，用户必须逐字母手打，非常低效。需要一键复制功能让用户快速获取占位符文本。

## What Changes

在 Settings.vue 的每个 field-chip 标签上新增独立的复制图标按钮（clipboard SVG 图标），位于 chip-label 和 chip-toggle 之间。点击复制按钮将 `${chip.name}` 复制到剪贴板，并显示 Tooltip "已复制"反馈（1.5 秒后自动消失）。

现有交互不变：点击 chip-label 仍触发添加/移除映射，点击 chip-toggle 仍切换启用/禁用。零新依赖，使用 `navigator.clipboard.writeText()` API。

## Capabilities

### New Capabilities
- `chip-copy-to-clipboard`: field-chip 标签上的复制按钮功能，包括复制到剪贴板和 Tooltip 反馈

### Modified Capabilities
（无现有 capability 的需求变更）

## Impact

- **前端代码**：`frontend/src/views/Settings.vue` — 模板新增复制按钮元素，script 新增复制逻辑和 Tooltip 状态管理，style 新增复制按钮和 Tooltip 样式
- **无后端变更**
- **无 API 变更**
- **无新依赖**
