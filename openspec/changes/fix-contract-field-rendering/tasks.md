## 1. 解析 fieldMap 和降级逻辑

- [x] 1.1 新增 `parsedFieldMap` computed：解析 `selectedTemplate.value.fieldMap` JSON，返回 `Record<string, string>`，解析失败返回空对象
- [x] 1.2 修改 `activeFieldsList` computed：优先从 `parsedFieldMap` 获取字段键列表，fallback 到 `activeFields` 的键列表
- [x] 1.3 修改 `getFieldLabel` 函数：优先从 `parsedFieldMap[key]` 读取标签，fallback 到 `fieldLabels[key]`，再 fallback 到 `key`

## 2. 自定义字段 v-model 管理

- [x] 2.1 新增 `customFieldValues = ref<Record<string, string>>({})` 状态
- [x] 2.2 在模板中为自定义字段（else 分支）绑定 `v-model="customFieldValues[key]"`
- [x] 2.3 在 `resetAll` 中清空 `customFieldValues`

## 3. 表单提交 payload 适配

- [x] 3.1 在 `createContract` 中将 `customFieldValues` 合并到 payload
- [x] 3.2 检查后端 contractApi.create 是否支持接收自定义字段，记录结果

## 4. 验证

- [x] 4.1 运行 `vue-tsc --noEmit` 类型检查通过
- [x] 4.2 运行 `npm run build` 构建通过
