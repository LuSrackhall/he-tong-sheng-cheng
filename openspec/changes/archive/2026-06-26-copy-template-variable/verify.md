## Verification Summary

**Change:** copy-template-variable

| 维度 | 状态 |
|------|------|
| 完整性 | 8/8 tasks 完成，3 个 spec 需求全覆盖 |
| 正确性 | 3/3 需求已实现并通过验证 |
| 一致性 | 实现遵循 design.md 所有决策 |

### Key Changes
- `frontend/src/views/Settings.vue` (script): 新增 `recentlyCopied` ref + `copyToClipboard` 函数
- `frontend/src/views/Settings.vue` (template): 每个 field-chip 新增复制按钮（SVG 图标）+ Tooltip 元素
- `frontend/src/views/Settings.vue` (style): `.chip-copy` 按钮样式 + `.copy-tooltip` Tooltip 样式 + `@keyframes copy-fade` 动画

### 需求覆盖
- [x] 复制字段占位符到剪贴板（预置字段 + 自定义字段）
- [x] 复制后即时反馈（Tooltip "已复制" + 1.5 秒自动消失）
- [x] 复制按钮不影响现有交互（@click.stop 事件隔离）

### Issues
无
