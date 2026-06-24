## 1. 后端：添加 signingDate 字段映射

- [x] 1.1 在 `buildReplaceValues` 函数中添加 `values["signingDate"] = contract.CreatedAt.Format("2006-01-02")`
- [x] 1.2 在 `PreviewTemplate` handler 的 `fields` 列表中添加 `signingDate` 条目（Key: "signingDate", Label: "签订日期", Required: false）
- [x] 1.3 在 `PreviewTemplate` handler 的 `builtinKeys` 集合中添加 `"signingDate": true`

## 2. 前端：添加 signingDate 预置字段

- [x] 2.1 在 `presetFieldGroups` 的"合同类"分组的 `fields` 数组中添加 `'signingDate'`
- [x] 2.2 在 `presetFieldLabels` 中添加 `signingDate: '签订日期'`
