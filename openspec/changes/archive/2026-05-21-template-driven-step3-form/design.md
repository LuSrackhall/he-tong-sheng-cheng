## Context

当前签合同流程 Step 3 的表单字段完全硬编码（开始日期、结束日期、月租金、年租金、应收总额、押金、备注 + 条件性 contractId），与 Settings 中模板的 fieldMap/activeFields 配置完全脱节。activeFields 当前为 `string[]` 格式，启用字段同时承担"参与渲染"和"参与 Word 校验"两个职责，无法独立控制。

## Goals / Non-Goals

**Goals:**
- Step 3 表单字段由模板 activeFields 完全驱动，模板启用什么就展示什么
- 字段按值来源分类渲染：用户填写型（可编辑）、系统自动型（只读）、资产/租户选择型（只读）
- activeFields 从 `string[]` 升级为 `Record<string, boolean>`，`false` = 启用但跳过 Word 校验
- 年租金 `yearlyRent` 作为独立预设字段，与月租金联动换算
- Settings 摘要区 chip 增加校验独立开关

**Non-Goals:**
- 不修改 Contract 数据库表结构
- 不新增自定义字段的动态值存储（`buildReplaceValues` 已支持 `fieldMap` 中任意 key 填充空字符串）
- 不改变催缴、收租等其他模块的表单展示

## Decisions

### D1: activeFields 格式升级为 `Record<string, boolean>`

**选择**: `{"startDate": true, "yearlyRent": false}`  
**原因**: 最简洁的结构，一个字段同时表达"是否启用"和"是否校验"。key 存在 = 启用，value = 是否校验。  
**替代方案**: 独立的 `excludeValidation: string[]` 数组——需要维护两份配置的同步，容易不一致。

向后兼容：`parseActiveFields` 检测到 `[]interface{}` 时自动转换为 `map[string]bool`，所有 value 默认为 `true`。

### D2: 字段分类驱动表单渲染

Step 3 表单根据字段的"值来源"决定渲染模式：

| 来源 | 字段 | 渲染模式 |
|------|------|----------|
| 系统自动 | contractId, totalReceivable, totalReceived, status, today | 只读 input，灰色背景 |
| 资产/租户 | assetName, assetType, assetDescription, tenantName, tenantIDCard, tenantPhone | 只读 input，灰色背景 |
| 用户填写 | startDate, endDate, monthlyRent, yearlyRent, deposit, notes | 可编辑 input |
| 自定义 | fieldMap 中非预置的任何字段 | 可编辑 input |

字段分类在前端通过常量映射维护，不依赖后端接口。

### D3: 月/年租金联动换算

- 表单中两者都存在（来自 activeFields）时，显示联动开关（默认开启）
- 联动开启：修改月租金 → 年租金 = 月租金 × 12；修改年租金 → 月租金 = 年租金 / 12
- 联动关闭：两者独立编辑
- Word 渲染时两者独立输出各自的值

### D4: 校验开关 UI 位置

校验开关放在"已启用字段"摘要区 chip 上（格式 `[${fieldName} ✓校验]`），点击切换为 `[${fieldName} ✗不校验]`。不在映射区 chip 上增加按钮以避免拥挤。

### D5: JSON 编辑器与 Chip 双向同步

- Chip 点击校验开关 → 更新 `activeFields[key]`，同步更新 JSON 编辑器 display（不改变 fieldMap 文本）
- JSON 编辑器手动修改 → `saveMapping` 时重新解析 activeFields 并提取校验状态
- 当前 `//` 注释机制不变：注释行 = 不启用 = 不在 activeFields 中

## Risks / Trade-offs

- [数据不一致] 如果用户在 JSON 编辑器手动修改而未点保存就切换模板，编辑丢失 → 现有行为不变（已通过 `mapping[t.id]` 绑定处理）
- [activeFields 迁移] 旧 `[]string` 格式需在前端和后端同时兼容 → 在 parseActiveFields 中统一处理，新旧格式一视同仁
- [yearlyRent 和 monthlyRent 不同步] 用户关闭联动后可能填错 → 联动默认开启，降低出错概率

## Open Questions

- 无
