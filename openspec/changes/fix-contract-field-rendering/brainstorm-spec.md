## Context

合同模板的字段映射系统有两个维度：`fieldMap`（JSON，字段名 → 中文标签）和 `activeFields`（JSON，字段名 → true/false 校验开关）。这两个维度的设计意图是：

- **fieldMap**：定义合同中需要的所有字段及其在 UI 表单中显示的中文标签
- **activeFields**：控制每个字段是否需要在上传的 Word 文档中进行占位符校验

当前 `NewContract.vue` 存在 bug：UI 表单的字段列表完全依赖 `activeFields` 中为 `true` 的字段，导致管理员关闭某个字段的 Word 校验后，该字段也从合同创建表单中消失了。

## Goals / Non-Goals

**Goals:**
- 修复 activeFieldsList 计算逻辑：改为从 fieldMap 获取所有字段，而非仅 activeFields 中为 true 的字段
- 修复 getFieldLabel 优先级：优先使用 fieldMap 中用户自定义的中文标签
- 确保 fieldMap 解析失败时有合理的降级行为
- 保持 isTemplateUsable 不变（模板可用性仍依赖校验开关）
- 表单提交 payload 适配自定义字段

**Non-Goals:**
- 不修改 Settings.vue 中的 activeFields 管理逻辑（其设计是正确的）
- 不修改后端 API 或 Template 数据模型
- 不添加自定义字段的输入类型推断（如日期选择器、数字输入等——全部使用文本输入框）

## Decisions

### Decision 1: activeFieldsList 数据源

**选择**：从 `fieldMap` 解析字段列表，`activeFields` 仅用于模板可用性检查（isTemplateUsable）。

**理由**：fieldMap 定义了"合同有哪些字段"，activeFields 定义了"Word 中需要校验哪些字段"。这两个概念不应混淆。

**降级策略**：如果 fieldMap 解析失败或为空，fallback 到 activeFields 的键列表（兼容旧数据）。

### Decision 2: 标签优先级

**选择**：`parsedFieldMap[key]` > `fieldLabels[key]` > `key`

**理由**：用户在 mapping 中配置的中文标签是第一优先级，硬编码标签是兜底。

### Decision 3: 自定义字段输入组件

**选择**：不在已知分类中的字段，统一渲染为文本输入框。

**理由**：当前系统已有此 fallback（第 813-815 行的 `<template v-else>` 分支），保持一致。未来如需支持类型推断，可在 fieldMap 中扩展类型标记，但不在本次修复范围内。

### Decision 4: 必填逻辑

**选择**：mapping 中配置的字段均为必填项（系统自动生成的字段除外，这些是只读展示字段）。

**理由**：问题描述明确了此要求。`requiredFieldKeys` 数组用于模板可用性检查，不在表单层面强制——表单层面的必填校验由 `validateStep3` 处理。

### Decision 5: 表单提交 payload

**选择**：在 `createContract` 中将所有非 system-auto 的字段值收集到 payload 的 `customFields` 子对象中（或直接平铺到 payload）。

**影响范围**：已确认后端 handler 使用固定结构体（CreateContractRequest），不支持自定义字段。自定义字段值在 payload 中以 `customFields` 子对象传递，后端会忽略未知字段，不影响现有功能。

## Risks / Trade-offs

- **[旧数据兼容]** → fieldMap 为空时 fallback 到 activeFields，确保旧模板仍可用。降级标签使用硬编码 fieldLabels。
- **[自定义字段提交]** → 已确认后端 contractApi.create 不支持自定义字段（固定结构体）。自定义字段值包含在 payload 的 `customFields` 子对象中，后端 Gin 默认忽略未知 JSON 字段，不影响创建流程。自定义字段在 UI 中可填写，但值不会持久化。
- **[fieldMap JSON 格式异常]** → try-catch 捕获解析错误，返回空对象，UI 退化为 fallback 路径。
