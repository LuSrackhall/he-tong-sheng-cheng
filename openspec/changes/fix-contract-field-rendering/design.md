## Context

修复 `NewContract.vue` 中 Step 3 合同详情表单的字段渲染逻辑。涉及单一文件的前端修复，无架构变更。

## Goals / Non-Goals

**Goals:**
- Step 3 表单从 fieldMap 获取字段列表和标签
- 降级兼容无 fieldMap 的旧模板

**Non-Goals:**
- 不修改后端 API
- 不修改 Settings.vue

## Decisions

### 数据流变更

**当前数据流：**
```
selectedTemplate.activeFields → parseActiveFieldsArray → filter(true) → activeFieldsList → 模板渲染
```

**修复后数据流：**
```
selectedTemplate.fieldMap → JSON.parse → parsedFieldMap (Record<string, string>)
                                            ↓
                              Object.keys() → activeFieldsList → 模板渲染
                                            ↓
                              getFieldLabel(key) → parsedFieldMap[key] || fieldLabels[key] || key

降级路径（fieldMap 为空/解析失败）：
selectedTemplate.activeFields → parseActiveFieldsArray → Object.keys() → activeFieldsList
```

### 实现细节

#### 1. 新增 `parsedFieldMap` computed

```typescript
const parsedFieldMap = computed<Record<string, string>>(() => {
  if (!selectedTemplate.value?.fieldMap) return {}
  try {
    const parsed = JSON.parse(selectedTemplate.value.fieldMap)
    if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
      return parsed as Record<string, string>
    }
    return {}
  } catch {
    return {}
  }
})
```

#### 2. 修改 `activeFieldsList` computed

```typescript
const activeFieldsList = computed(() => {
  if (!selectedTemplate.value) return []
  // 优先从 fieldMap 获取字段列表
  const fm = parsedFieldMap.value
  if (Object.keys(fm).length > 0) return Object.keys(fm)
  // fallback: 从 activeFields 获取（兼容旧数据）
  const afMap = parseActiveFieldsArray(selectedTemplate.value.activeFields || '')
  return Object.keys(afMap)
})
```

#### 3. 修改 `getFieldLabel` 函数

```typescript
function getFieldLabel(key: string): string {
  return parsedFieldMap.value[key] || fieldLabels[key] || key
}
```

#### 4. 表单提交 payload 适配

在 `createContract` 中，将自定义字段值收集到 payload。需要：
- 新增一个 `customFieldValues` ref 对象，存储自定义字段的 v-model 值
- 在模板中为自定义字段绑定 v-model
- 提交时将自定义字段值合并到 payload

#### 5. 自定义字段的 v-model 管理

对于不在已知分类（SYSTEM_AUTO_FIELDS / ASSET_TENANT_FIELDS / USER_INPUT_FIELDS）中的字段：
- 新增 `customFieldValues = ref<Record<string, string>>({})`
- 模板中 `<input v-model="customFieldValues[key]">`
- 提交时合并到 payload

### isTemplateUsable 不变

`isTemplateUsable` 函数（第 55-59 行）继续使用 `activeFields` 检查必填字段是否在 Word 中配置了校验。这是正确的——模板可用性依赖于 Word 校验开关。

## Risks / Trade-offs

- **[旧数据兼容]** → fieldMap 为空时 fallback 到 activeFields。旧模板行为不变。
- **[自定义字段提交]** → 已确认后端使用固定结构体，不支持自定义字段。自定义字段值在 payload 的 `customFields` 子对象中传递，后端 Gin 忽略未知字段。
- **[v-model 状态管理]** → customFieldValues 需要在 resetAll 中清空。
