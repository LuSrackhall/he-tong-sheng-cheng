## 1. Backend: activeFields format upgrade

- [x] 1.1 Update `parseActiveFields` in `internal/transport/handler/template.go` to support both legacy `[]string` and new `Record<string, boolean>` format, returning `map[string]bool`
- [x] 1.2 Update `CreateTemplate` in `internal/transport/handler/contract.go` to initialize `activeFields` as `{}` (new default)
- [x] 1.3 Update `UpdateTemplateMapping` in `internal/transport/handler/contract.go` to auto-sync activeFields from fieldMap uncommented keys, preserving existing validation flags
- [x] 1.4 Update `UploadTemplate` validation in `internal/transport/handler/template.go` to only validate fields where `activeFields[key] === true`
- [x] 1.5 Update `UpdateTemplateMapping` re-validation in `internal/transport/handler/contract.go` to only validate fields where `activeFields[key] === true`
- [x] 1.6 Update `ExportContract` in `internal/transport/handler/template.go` to extract enabled field keys for required-fields check
- [x] 1.7 Update `Create` in `internal/transport/handler/contract.go` to extract enabled field keys for required-fields check

## 2. Frontend: activeFields format upgrade

- [x] 2.1 Update `parseActiveFieldsArray` in `Settings.vue` to return `Record<string, boolean>` from both legacy and new formats
- [x] 2.2 Update `isActive` in `Settings.vue` to read from `Record<string, boolean>` (key exists = active)
- [x] 2.3 Update `toggleActive` in `Settings.vue` to add/remove keys from the object instead of commenting/uncommenting
- [x] 2.4 Update `saveMapping` in `Settings.vue` to serialize activeFields as object and auto-sync with fieldMap uncommented keys
- [x] 2.5 Update `validateJson` in `Settings.vue` to read enabled keys from the new format
- [x] 2.6 Update `formatJson` in `Settings.vue` to reset activeFields to object format
- [x] 2.7 Update `isTemplateUsable` and `templateUnusableReason` in `Settings.vue` to read enabled keys from object
- [x] 2.8 Update corresponding functions in `NewContract.vue` (isTemplateUsable, parseActiveFieldsArray)

## 3. yearlyRent preset field

- [x] 3.1 Add `yearlyRent` to `presetFieldGroups` in `Settings.vue` (合同类 group)
- [x] 3.2 Add `yearlyRent` to `presetFieldLabels` in `Settings.vue` with label "年租金"
- [x] 3.3 Add `yearlyRent` to `buildReplaceValues` in `internal/transport/handler/template.go` (calculate from monthlyRent × 12 or user input)
- [x] 3.4 Add `yearlyRent` to default `FieldMap` in `CreateTemplate` handler

## 4. Step 3 form template-driven rendering

- [x] 4.1 Define field source classification constants in `NewContract.vue` (system-auto, asset-tenant, user-input sets)
- [x] 4.2 Create computed `activeFieldsList` that derives field list from selected template's activeFields
- [x] 4.3 Implement dynamic form field rendering: iterate activeFieldsList, render read-only or editable input based on source classification
- [x] 4.4 Remove all hardcoded form fields from Step 3 template (keep only the container and iteration logic)
- [x] 4.5 Implement read-only values for system-auto fields (contractId shows "将在创建后自动生成", others compute from data)
- [x] 4.6 Implement read-only values for asset/tenant fields (from selectedAsset/selectedTenant)
- [x] 4.7 Keep fallback: when no template selected, show startDate, endDate, monthlyRent as minimum required fields

## 5. monthlyRent/yearlyRent linked conversion in Step 3

- [x] 5.1 Add linked conversion logic: watch monthlyRent → update yearlyRent (×12), watch yearlyRent → update monthlyRent (/12)
- [x] 5.2 Add linkage toggle state (default on) and UI toggle button between the two fields
- [x] 5.3 When linkage toggled off, both fields become independently editable
- [x] 5.4 When linkage toggled back on, recalculate yearlyRent from current monthlyRent
- [x] 5.5 Only show linkage controls when both monthlyRent and yearlyRent are active

## 6. Per-field validation toggle in Settings summary

- [x] 6.1 Update "已启用字段" summary section to render chips with validation state indicator (`✓校验` / `✗不校验`)
- [x] 6.2 Add click handler on validation indicator to toggle `activeFields[key]` boolean value
- [x] 6.3 Sync validation toggle changes back to the activeFields data used in saveMapping
- [x] 6.4 Update chip styling to visually distinguish validate=true from validate=false (e.g., green vs gray)

## 7. Integration & verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `npm run build` passes
- [x] 7.3 Manual smoke test: create template, upload Word, configure fields with mixed validation toggles, create contract, export
