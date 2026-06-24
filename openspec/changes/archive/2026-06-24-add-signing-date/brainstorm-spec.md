# add-signing-date: 签订日期字段设计

## Context

合同支持 Word 模板动态生成（docx 占位符替换）。字段映射系统允许用户在 Word 模板中使用 `${字段名}` 格式的占位符，系统在生成时自动替换为实际数据。

当前系统中已有 `today` 字段（`time.Now()`），但每次下载时日期会变，不适合作为签订日期。数据库中 Contract 实体已有 `CreatedAt` 字段（合同创建时间），语义上等同于签订日期，但字段映射系统中没有暴露此字段。

## Goals / Non-Goals

**Goals:**
- 在字段映射系统中添加 `signingDate` 字段，映射到 `contract.CreatedAt`
- 用户可在 Word 模板中使用 `${signingDate}` 自动填充签订日期
- 日期格式与现有字段一致：`2006-01-02`

**Non-Goals:**
- 不添加用户可配置的日期格式（当前所有日期字段都是硬编码格式，保持一致）
- 不修改数据库 schema（CreatedAt 已由 GORM 自动管理）
- 不新增 API 端点

## Decisions

### 1. signingDate 映射到 contract.CreatedAt

**选择：** `contract.CreatedAt.Format("2006-01-02")`

**理由：** 合同创建时间即签订日期。CreatedAt 由 GORM 在记录创建时自动设置，之后不会改变，适合作为签订日期。

### 2. 字段归属"合同类"分组

**选择：** 将 signingDate 放入前端 `presetFieldGroups` 的"合同类"分组。

**理由：** 签订日期是合同的核心属性，与 startDate/endDate/contractId 等同属合同类。

### 3. 不设为必填字段

**选择：** signingDate 不加入 requiredFields 列表。

**理由：** 并非所有合同模板都需要签订日期，用户按需启用即可。

## Risks / Trade-offs

- **[风险] 合同创建时间 ≠ 实际签订时间** → 当前系统无独立的签订日期字段，CreatedAt 是最佳近似值。如未来需要精确签订时间，可添加独立字段。
