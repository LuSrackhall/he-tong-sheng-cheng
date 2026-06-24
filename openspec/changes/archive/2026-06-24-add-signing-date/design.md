## Context

在资产租赁系统的合同模板字段映射机制中，新增 `signingDate` 内置字段。该字段映射到 `contract.CreatedAt`（数据库创建时间），格式 `2006-01-02`。

现有架构：后端 `buildReplaceValues` 构建 `map[string]string` 供 docx 占位符替换；前端 `presetFieldGroups` 定义预置字段分组供用户选择。新增字段只需在这两处各加一行。

## Goals / Non-Goals

**Goals:**
- 用户可在 Word 模板中使用 `${signingDate}`，生成时自动替换为合同创建日期
- 字段在前端"合同类"分组中可见可选

**Non-Goals:**
- 不引入日期格式配置机制（保持与现有字段一致的硬编码格式）
- 不修改数据库 schema

## Decisions

### 1. signingDate 数据源 = contract.CreatedAt

直接使用 GORM 自动设置的 `CreatedAt`，无需新增字段或数据库迁移。`CreatedAt` 在记录创建后不会改变，适合代表签订日期。

### 2. 放入"合同类"预置分组

与 `startDate`、`endDate`、`contractId` 等同属合同信息，归入"合同类"分组最自然。

### 3. 不设为必填字段

签订日期不是所有模板的必选项，用户按需启用。

## Risks / Trade-offs

- **[语义偏差]** 合同创建时间 ≠ 实际签字时间 → 可接受，当前系统无独立签订日期字段，CreatedAt 是最佳近似。
