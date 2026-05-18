# Fix Template Validation on Mapping Update — Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development
> to implement this plan task-by-task.

**Goal:** Fix docx placeholder validation not working on upload, and add re-validation on mapping update with a Validated status gate for export.

**Architecture:** Two root causes — `io.Reader.Read` doesn't guarantee full read (validation silently broken), and `UpdateTemplateMapping` never re-validates after saving. Fix: use `io.ReadAll`, add `Validated bool` to Template domain model, re-validate in `UpdateTemplateMapping`, gate `ExportContract` on Validated.

**Tech Stack:** Go + Gin + GORM (dual SQLite/PostgreSQL), Vue 3 + TypeScript + Axios

---

### Task 1: Domain — Add Validated field

**Files:**
- Modify: `internal/domain/template.go`

- [ ] **Step 1: Add Validated field to Template struct**

```go
type Template struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	FilePath     string    `json:"filePath" gorm:"default:''"`
	FieldMap     string    `json:"fieldMap,omitempty" gorm:"type:text"`
	ActiveFields string    `json:"activeFields,omitempty" gorm:"type:text"`
	Validated    bool      `json:"validated" gorm:"default:false"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: PASS (new field, no callers yet)

- [ ] **Step 3: Commit**

```bash
git add internal/domain/template.go
git commit -m "feat: add Validated field to Template domain model"
```

---

### Task 2: Fix UploadTemplate — io.ReadAll + Validated

**Files:**
- Modify: `internal/transport/handler/template.go:53`

- [ ] **Step 1: Replace src.Read with io.ReadAll**

Change `template.go:53` from:
```go
fileData := make([]byte, file.Size)
if _, err := src.Read(fileData); err != nil {
```
To:
```go
fileData, err := io.ReadAll(src)
if err != nil {
```

Also remove the `"io"` import is already used — check if `"io"` is already imported; if not, add it to the import block.

- [ ] **Step 2: Set tpl.Validated based on validation result**

After the validation block (lines 59-74), update the logic:

```go
// Validate placeholders against active fields
activeFields := parseActiveFields(tpl.ActiveFields)
if len(activeFields) > 0 {
	missing, err := docx.ValidatePlaceholders(fileData, activeFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse docx: " + err.Error()})
		return
	}
	if len(missing) > 0 {
		// Set validated=false before returning error
		tpl.Validated = false
		h.templateRepo.Update(tpl)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Word 文件缺少以下已启用的占位符",
			"missingFields": missing,
		})
		return
	}
}
// Validation passed (or no active fields to check)
tpl.Validated = true
```

And after saving the file (just before `c.JSON(http.StatusOK, tpl)`):
```go
// tpl.FilePath already set above, and tpl.Validated is already true
if err := h.templateRepo.Update(tpl); err != nil {
```

Wait — looking at the existing code more carefully, `tpl.FilePath` is set and `h.templateRepo.Update(tpl)` is called after. We need to set Validated before that Update call. Let me re-read the existing flow.

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/transport/handler/template.go
git commit -m "fix: use io.ReadAll for complete file read and set Validated on upload"
```

---

### Task 3: Backend — Re-validation in UpdateTemplateMapping

**Files:**
- Modify: `internal/transport/handler/contract.go:209-241`

- [ ] **Step 1: Add re-validation logic after saving mapping**

In `UpdateTemplateMapping`, after `h.templateRepo.Update(tpl)`, add:

```go
// Re-validate existing Word file if present
if tpl.FilePath != "" {
	fileData, err := os.ReadFile(tpl.FilePath)
	if err != nil {
		// File missing on disk — treat as not validated
		tpl.Validated = false
	} else {
		activeFields := parseActiveFields(tpl.ActiveFields)
		if len(activeFields) > 0 {
			missing, err := docx.ValidatePlaceholders(fileData, activeFields)
			if err != nil {
				tpl.Validated = false
			} else {
				tpl.Validated = len(missing) == 0
			}
		} else {
			// No active fields — consider validated
			tpl.Validated = true
		}
	}
	h.templateRepo.Update(tpl)
}
```

Add these imports to contract.go if not already present:
```go
"asset-leasing-system/internal/docx"
"os"
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/transport/handler/contract.go
git commit -m "feat: re-validate uploaded Word file on mapping update"
```

---

### Task 4: Backend — Export gate on Validated

**Files:**
- Modify: `internal/transport/handler/template.go:97-159`

- [ ] **Step 1: Add Validated check in ExportContract**

In `ExportContract`, after the `tpl.FilePath == ""` check (line 122-125), add:

```go
if !tpl.Validated {
	c.JSON(http.StatusConflict, gin.H{"error": "模板校验未通过，请先上传符合要求的 Word 文件"})
	return
}
```

This goes right before `templateData, err := os.ReadFile(tpl.FilePath)`.

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/transport/handler/template.go
git commit -m "feat: block contract export when template validation fails"
```

---

### Task 5: Frontend — Template type + API

**Files:**
- Modify: `frontend/src/api/index.ts`

- [ ] **Step 1: Add validated to Template interface**

```typescript
export interface Template {
  id: number
  name: string
  filePath: string
  fieldMap?: string
  activeFields?: string
  validated: boolean
  createdAt: string
}
```

- [ ] **Step 2: Verify TypeScript compilation**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1`
Expected: May have errors in Settings.vue (handled next task), but the API file should be fine.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/index.ts
git commit -m "feat: add validated field to Template type"
```

---

### Task 6: Frontend — Validation status in Settings

**Files:**
- Modify: `frontend/src/views/Settings.vue`

- [ ] **Step 1: Display validation status badge**

In the template list section, add a status indicator next to each template name. Find the template name display and add:

```vue
<span v-if="t.filePath && t.validated" class="badge-validated">校验通过</span>
<span v-else-if="t.filePath && !t.validated" class="badge-not-validated">校验未通过</span>
```

Add CSS:
```css
.badge-validated { color: #22c55e; font-size: 12px; margin-left: 8px; }
.badge-not-validated { color: #ef4444; font-size: 12px; margin-left: 8px; }
```

- [ ] **Step 2: Show notification after mapping save**

In the `saveMapping` function, after successful save, check `res.data.validated`:

```typescript
const res = await templateApi.updateMapping(t.id, fieldMap, activeFields)
// After successful save:
if (res.data.filePath) {
  if (res.data.validated) {
    // Success notification
    ElMessage.success('映射已保存，Word 文件校验通过')
  } else {
    ElMessage.warning('映射已保存，但 Word 文件校验未通过，请重新上传符合要求的 Word 文件')
  }
} else {
  ElMessage.success('映射已保存')
}
```

- [ ] **Step 3: Verify frontend build**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 && npm run build 2>&1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/Settings.vue
git commit -m "feat: display template validation status in Settings"
```

---

### Task 7: Frontend — Export error handling

**Files:**
- Modify: `frontend/src/views/` (the contract export view)

- [ ] **Step 1: Find the export function and handle 409**

Locate the contract export button handler. Add 409 handling:

```typescript
} catch (err: any) {
  if (err.response?.status === 409) {
    ElMessage.error('模板校验未通过，请先在设置中重新上传符合要求的 Word 文件')
  } else {
    ElMessage.error('导出失败')
  }
}
```

- [ ] **Step 2: Verify frontend build**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 && npm run build 2>&1`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/
git commit -m "feat: handle template validation error in export flow"
```

---

### Task 8: Full build verification

- [ ] **Step 1: Backend build**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 2: Frontend type check + build**

Run: `cd frontend && npx vue-tsc --noEmit 2>&1 && npm run build 2>&1`
Expected: PASS

- [ ] **Step 3: Final commit (if any cleanups needed)**

```bash
git add -A
git commit -m "chore: final verification fixes"
```
