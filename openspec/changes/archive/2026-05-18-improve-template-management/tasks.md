## 1. Backend — Template Deletion

- [x] 1.1 Add `Delete` and `IsUsedByContract` methods to `TemplateRepo` interface in `internal/domain/template.go`
- [x] 1.2 Implement `Delete` and `IsUsedByContract` in the repository layer (GORM)
- [x] 1.3 Add `DeleteTemplate` handler in `internal/transport/handler/template.go` with contract reference check (409 if used)
- [x] 1.4 Register `DELETE /api/templates/:id` route in `cmd/server/main.go`

## 2. Backend — Validation Logic (All Active Fields Must Be Validated)

- [x] 2.1 Verify `UploadTemplate` in `template.go` already validates all `activeFields` against Word placeholders (no change needed per current code)

## 3. Frontend — Template Deletion

- [x] 3.1 Add `deleteTemplate` API function in `frontend/src/api/index.ts`
- [x] 3.2 Add delete button to each template row in `Settings.vue` with confirmation dialog
- [x] 3.3 Handle 409 error (template in use) with user-friendly notification

## 4. Frontend — Custom Field Addition

- [x] 4.1 Add "添加自定义字段" button and modal form (field name + display label inputs) in `Settings.vue`
- [x] 4.2 Add client-side validation: reject empty field name/label, reject duplicate field names
- [x] 4.3 Append new custom field to `fieldMap` JSON and save via existing mapping API

## 5. Frontend — Auto-Enable on Field Add

- [x] 5.1 Modify `insertFieldPlaceholder` to automatically add the field to `activeSet` after insertion
- [x] 5.2 Ensure custom field creation also auto-adds to `activeFields`

## 6. Frontend — Field Label Display

- [x] 6.1 Update preset field tag display from `${fieldName}` to `${fieldName} → label` using `fieldMap` values
- [x] 6.2 Handle edge case: field without a label falls back to showing only `${fieldName}`
