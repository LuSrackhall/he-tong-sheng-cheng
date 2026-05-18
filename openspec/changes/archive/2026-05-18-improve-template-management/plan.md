# 模板管理功能改进 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现模板删除、自定义字段映射、自动启用、字段标签展示四项改进。

**Architecture:** 后端 Go+Gin，在 TemplateRepo 接口新增 Delete/IsUsedByContract 方法，handler 新增 DeleteTemplate 端点；前端 Vue 3，修改 Settings.vue 添加删除按钮、自定义字段弹窗、自动启用逻辑和标签展示。无需数据库迁移。

**Tech Stack:** Go + Gin + GORM, Vue 3 + TypeScript

---

## Task 1: Backend — TemplateRepo 接口和实现

**Files:**
- Modify: `internal/domain/repo.go:49-54`
- Modify: `internal/repository/sqlite/repos.go:64-85`
- Modify: `internal/repository/postgres/contract.go:137-158`

- [ ] **Step 1: 在 TemplateRepo 接口中添加 Delete 和 IsUsedByContract 方法**

编辑 `internal/domain/repo.go`，在 TemplateRepo 接口中追加两个方法：

```go
type TemplateRepo interface {
	Create(t *Template) error
	GetByID(id uint) (*Template, error)
	List() ([]Template, error)
	Update(t *Template) error
	Delete(id uint) error
	IsUsedByContract(id uint) (bool, error)
}
```

- [ ] **Step 2: 在 SQLite 实现中添加两个方法**

编辑 `internal/repository/sqlite/repos.go`，在 TemplateRepo 的 Update 方法后追加：

```go
func (r *TemplateRepo) Delete(id uint) error {
	return r.db.Delete(&domain.Template{}, id).Error
}

func (r *TemplateRepo) IsUsedByContract(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Contract{}).Where("template_id = ?", id).Count(&count).Error
	return count > 0, err
}
```

- [ ] **Step 3: 在 Postgres 实现中添加两个方法**

编辑 `internal/repository/postgres/contract.go`，在 TemplateRepo 的 Update 方法后追加：

```go
func (r *TemplateRepo) Delete(id uint) error {
	return r.db.Delete(&domain.Template{}, id).Error
}

func (r *TemplateRepo) IsUsedByContract(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Contract{}).Where("template_id = ?", id).Count(&count).Error
	return count > 0, err
}
```

- [ ] **Step 4: 验证编译通过**

```bash
cd /Users/srackhalllu/Desktop/资源管理器/safe/he-tong-sheng-cheng && go build ./...
```

Expected: 编译成功，无错误。

- [ ] **Step 5: Commit**

```bash
git add internal/domain/repo.go internal/repository/sqlite/repos.go internal/repository/postgres/contract.go
git commit -m "feat: add Delete and IsUsedByContract to TemplateRepo interface and implementations"
```

---

## Task 2: Backend — DeleteTemplate Handler 和路由

**Files:**
- Modify: `internal/transport/handler/template.go` (追加 handler)
- Modify: `cmd/server/main.go:75-78` (注册路由)

- [ ] **Step 1: 添加 DeleteTemplate handler**

编辑 `internal/transport/handler/template.go`，在文件末尾追加：

```go
// DeleteTemplate handles DELETE /api/templates/:id
func (h *ContractHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	_, err = h.templateRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	used, err := h.templateRepo.IsUsedByContract(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check template usage"})
		return
	}
	if used {
		c.JSON(http.StatusConflict, gin.H{"error": "该模板已被合同引用，无法删除"})
		return
	}

	if err := h.templateRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模板已删除"})
}
```

- [ ] **Step 2: 注册 DELETE 路由**

编辑 `cmd/server/main.go`，在 templates 路由组中添加 DELETE 路由：

```go
protected.GET("/templates", contractH.ListTemplates)
protected.POST("/templates", contractH.CreateTemplate)
protected.PATCH("/templates/:id", contractH.UpdateTemplateMapping)
protected.POST("/templates/:id/upload", contractH.UploadTemplate)
protected.DELETE("/templates/:id", contractH.DeleteTemplate)
```

- [ ] **Step 3: 验证编译通过**

```bash
cd /Users/srackhalllu/Desktop/资源管理器/safe/he-tong-sheng-cheng && go build ./...
```

Expected: 编译成功。

- [ ] **Step 4: Commit**

```bash
git add internal/transport/handler/template.go cmd/server/main.go
git commit -m "feat: add DELETE /api/templates/:id endpoint with contract reference check"
```

---

## Task 3: Frontend — deleteTemplate API

**Files:**
- Modify: `frontend/src/api/index.ts:63-78`

- [ ] **Step 1: 在 templateApi 中添加 deleteTemplate 方法**

编辑 `frontend/src/api/index.ts`，在 `templateApi` 对象中添加：

```typescript
export const templateApi = {
  list: () => api.get<{ data: Template[] } | Template[]>('/templates'),
  create: (name: string) => api.post<Template>('/templates', { name }),
  updateMapping: (id: number, fieldMap: string, activeFields: string) =>
    api.patch<Template>(`/templates/${id}`, { fieldMap, activeFields }),
  delete: (id: number) => api.delete<{ message: string }>(`/templates/${id}`),
  uploadFile: (id: number, file: File, onProgress?: (pct: number) => void) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post<Template>(`/templates/${id}/upload`, fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e: any) => {
        if (e.total && onProgress) onProgress(Math.round((e.loaded * 100) / e.total))
      },
    })
  },
}
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
cd /Users/srackhalllu/Desktop/资源管理器/safe/he-tong-sheng-cheng/frontend && npx vue-tsc --noEmit 2>&1 | head -20
```

Expected: 无新增类型错误。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/index.ts
git commit -m "feat: add deleteTemplate API method"
```

---

## Task 4: Frontend — Settings.vue 删除按钮

**Files:**
- Modify: `frontend/src/views/Settings.vue`

- [ ] **Step 1: 在 card-header 中添加删除按钮**

编辑 `frontend/src/views/Settings.vue`，在 `<script setup>` 部分添加 `deleting` ref 和 `deleteTemplate` 函数：

```typescript
const deleting = ref<Record<number, boolean>>({})

async function deleteTemplate(t: Template) {
  if (!confirm(`确定要删除模板「${t.name}」吗？此操作不可撤销。`)) return
  deleting.value[t.id] = true
  try {
    await templateApi.delete(t.id)
    flash('模板已删除')
    await fetchTemplates()
  } catch (e: any) {
    const status = e.response?.status
    if (status === 409) {
      flashError('该模板已被合同引用，无法删除')
    } else {
      flashError(e.response?.data?.error || '删除失败')
    }
  } finally {
    deleting.value[t.id] = false
  }
}
```

- [ ] **Step 2: 在模板卡片头部渲染删除按钮**

在 `card-title-row` div 内的 `file-badge` 后添加删除按钮：

```html
<div class="card-title-row">
  <span class="template-name">{{ t.name }}</span>
  <span :class="['file-badge', hasFile(t) ? 'file-ok' : 'file-missing']">
    {{ hasFile(t) ? '文件已上传' : '未上传文件' }}
  </span>
  <button
    class="btn-delete-template"
    :disabled="deleting[t.id]"
    @click="deleteTemplate(t)"
    :title="deleting[t.id] ? '删除中...' : '删除模板'"
  >
    {{ deleting[t.id] ? '...' : '✕' }}
  </button>
</div>
```

- [ ] **Step 3: 添加删除按钮样式**

在 `<style scoped>` 末尾追加：

```css
.btn-delete-template {
  margin-left: auto;
  width: 26px; height: 26px;
  display: inline-flex; align-items: center; justify-content: center;
  padding: 0; border: 1px solid transparent; border-radius: 50%;
  background: transparent; font-size: 0.75rem; color: var(--color-text-tertiary);
  cursor: pointer; transition: all var(--transition-fast);
}
.btn-delete-template:hover {
  border-color: var(--color-danger);
  color: var(--color-danger);
  background: rgba(255,59,48,0.06);
}
.btn-delete-template:disabled {
  opacity: 0.5; cursor: not-allowed;
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/Settings.vue
git commit -m "feat: add delete button to template cards with confirmation and 409 handling"
```

---

## Task 5: Frontend — 自定义字段弹窗

**Files:**
- Modify: `frontend/src/views/Settings.vue`

- [ ] **Step 1: 添加自定义字段弹窗状态和逻辑**

在 `<script setup>` 中添加：

```typescript
// Custom field modal state (per template)
const showCustomField = ref<Record<number, boolean>>({})
const customFieldName = ref<Record<number, string>>({})
const customFieldLabel = ref<Record<number, string>>({})
const customFieldError = ref<Record<number, string>>({})

function openCustomFieldModal(templateId: number) {
  customFieldName.value[templateId] = ''
  customFieldLabel.value[templateId] = ''
  customFieldError.value[templateId] = ''
  showCustomField.value[templateId] = true
}

function addCustomField(templateId: number) {
  const name = (customFieldName.value[templateId] || '').trim()
  const label = (customFieldLabel.value[templateId] || '').trim()

  if (!name) {
    customFieldError.value[templateId] = '字段名不能为空'
    return
  }
  if (!label) {
    customFieldError.value[templateId] = '显示标签不能为空'
    return
  }

  // Check for duplicate field names (preset + custom)
  const allPresetFields = presetFieldGroups.flatMap(g => g.fields)
  const existingMapKeys = getMapKeys(templateId)
  if (allPresetFields.includes(name) || existingMapKeys.includes(name)) {
    customFieldError.value[templateId] = `字段名 "${name}" 已存在`
    return
  }

  // Add to fieldMap
  const current = mapping.value[templateId] || '{}'
  try {
    const obj = parseFieldMap(current)
    obj[name] = label
    mapping.value[templateId] = JSON.stringify(obj, null, 2)
  } catch {
    mapping.value[templateId] = JSON.stringify({ [name]: label }, null, 2)
  }

  // Auto-enable: add to activeSet
  if (!activeSet.value[templateId]) {
    activeSet.value[templateId] = new Set()
  }
  const s = activeSet.value[templateId]
  s.add(name)
  activeSet.value[templateId] = new Set(s)

  showCustomField.value[templateId] = false
  jsonErrors.value[templateId] = ''
}
```

- [ ] **Step 2: 在预设字段区域后添加"添加自定义字段"按钮**

在 `preset-fields` div 闭合标签 `</div>` 后（JSON editor textarea 前）添加按钮：

```html
<button
  class="btn-add-custom"
  type="button"
  @click="openCustomFieldModal(t.id)"
>
  + 添加自定义字段
</button>
```

- [ ] **Step 3: 在 template-card 内部添加自定义字段弹窗**

在 template-card 的闭合 `</div>` 前（`</details>` 后）添加：

```html
<!-- Custom field modal -->
<div v-if="showCustomField[t.id]" class="modal-overlay" @click.self="showCustomField[t.id] = false">
  <div class="modal-content" style="max-width: 380px;">
    <h3>添加自定义字段</h3>
    <div class="form-group">
      <label class="label">字段名（占位符）</label>
      <input
        class="input"
        v-model="customFieldName[t.id]"
        placeholder="例如：customField"
        @keyup.enter="addCustomField(t.id)"
      />
    </div>
    <div class="form-group">
      <label class="label">显示标签</label>
      <input
        class="input"
        v-model="customFieldLabel[t.id]"
        placeholder="例如：自定义字段"
        @keyup.enter="addCustomField(t.id)"
      />
    </div>
    <div v-if="customFieldError[t.id]" class="json-error" style="margin-bottom: 8px;">{{ customFieldError[t.id] }}</div>
    <div class="modal-actions">
      <button class="btn btn-secondary" @click="showCustomField[t.id] = false">取消</button>
      <button class="btn btn-primary" @click="addCustomField(t.id)">添加</button>
    </div>
  </div>
</div>
```

- [ ] **Step 4: 添加按钮样式**

在 `<style scoped>` 末尾追加：

```css
.btn-add-custom {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; margin-bottom: 12px;
  border: 1px dashed var(--color-border); border-radius: var(--radius-sm);
  background: transparent; color: var(--color-text-secondary);
  font-size: 0.75rem; cursor: pointer;
  transition: all var(--transition-fast);
}
.btn-add-custom:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background: rgba(0,122,255,0.04);
}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/Settings.vue
git commit -m "feat: add custom field modal with validation and auto-enable"
```

---

## Task 6: Frontend — 自动启用逻辑

**Files:**
- Modify: `frontend/src/views/Settings.vue`

- [ ] **Step 1: 修改 insertFieldPlaceholder 自动启用**

修改 `insertFieldPlaceholder` 函数，在添加字段到 fieldMap 后自动加入 activeSet：

```typescript
function insertFieldPlaceholder(templateId: number, fieldName: string) {
  const current = mapping.value[templateId] || '{}'
  try {
    const obj = parseFieldMap(current)
    if (!obj[fieldName]) {
      obj[fieldName] = fieldName
      mapping.value[templateId] = JSON.stringify(obj, null, 2)
    }
  } catch {
    mapping.value[templateId] = JSON.stringify({ [fieldName]: fieldName }, null, 2)
  }

  // Auto-enable: add to activeSet
  if (!activeSet.value[templateId]) {
    activeSet.value[templateId] = new Set()
  }
  const s = activeSet.value[templateId]
  if (!s.has(fieldName)) {
    s.add(fieldName)
    activeSet.value[templateId] = new Set(s)
  }

  jsonErrors.value[templateId] = ''
}
```

- [ ] **Step 2: 更新 section-hint 文案**

将字段映射配置区的 section-hint 文案改为：

```html
<span class="section-hint">
  点击标签添加映射并自动启用，关闭开关则禁用该字段（不参与替换和校验）
</span>
```

- [ ] **Step 3: 更新 Word 说明区文案**

将 guide-content 中关于校验的说明改为：

```html
<p>上传文件时，系统会校验<strong>所有已启用的</strong>字段是否都在 Word 中存在对应占位符。关闭开关的字段不会被校验。</p>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/Settings.vue
git commit -m "feat: auto-enable field on add, update UI hints to reflect validation behavior"
```

---

## Task 7: Frontend — 字段标签展示

**Files:**
- Modify: `frontend/src/views/Settings.vue`

- [ ] **Step 1: 修改 chip-label 显示标签**

将预设字段的 chip-label 从只显示占位符改为显示 `占位符 → 标签`：

```html
<button
  class="chip-label"
  type="button"
  @click="insertFieldPlaceholder(t.id, field)"
  :title="'添加到映射'">
  ${{ '{' + field + '}' }}
  <span class="chip-label-text">→ {{ parseFieldMap(mapping[t.id] || '{}')[field] || field }}</span>
</button>
```

- [ ] **Step 2: 添加标签文字样式**

在 `<style scoped>` 中添加：

```css
.chip-label-text {
  font-family: inherit;
  font-size: 0.7rem;
  color: var(--color-text-secondary);
  margin-left: 2px;
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/Settings.vue
git commit -m "feat: display field label alongside placeholder in preset chips"
```

---

## Task 8: 端到端验证

- [ ] **Step 1: 启动后端**

```bash
cd /Users/srackhalllu/Desktop/资源管理器/safe/he-tong-sheng-cheng && go run ./cmd/server &
sleep 2
```

- [ ] **Step 2: 测试 DELETE 端点 — 删除不存在的模板**

```bash
curl -s -o /dev/null -w "%{http_code}" -X DELETE http://localhost:8080/api/templates/99999 -H "Authorization: Bearer <token>"
```

Expected: 404

- [ ] **Step 3: 测试 DELETE 端点 — 删除被引用的模板（若存在）**

先创建一个模板和一个引用它的合同，再尝试删除：

```bash
# 预期 409
curl -s -X DELETE http://localhost:8080/api/templates/<used-id> -H "Authorization: Bearer <token>"
```

Expected: 409, `{"error":"该模板已被合同引用，无法删除"}`

- [ ] **Step 4: 测试 DELETE 端点 — 成功删除未使用的模板**

```bash
# 创建新模板
curl -s -X POST http://localhost:8080/api/templates -H "Authorization: Bearer <token>" -H "Content-Type: application/json" -d '{"name":"test-delete"}'
# 记下返回的 id，然后删除
curl -s -X DELETE http://localhost:8080/api/templates/<id> -H "Authorization: Bearer <token>"
```

Expected: 200, `{"message":"模板已删除"}`

- [ ] **Step 5: 构建前端并验证**

```bash
cd /Users/srackhalllu/Desktop/资源管理器/safe/he-tong-sheng-cheng/frontend && npm run build
```

Expected: 构建成功。

- [ ] **Step 6: 清理后台进程**

```bash
kill %1 2>/dev/null
```

- [ ] **Step 7: Commit (如有前端构建产物变更)**

```bash
git add -A
git commit -m "chore: frontend build output"
```
