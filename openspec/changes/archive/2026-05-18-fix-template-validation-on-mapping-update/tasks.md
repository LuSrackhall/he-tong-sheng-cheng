## 1. Backend — Domain

- [ ] 1.1 Add `Validated bool` field to `Template` struct in `internal/domain/template.go`

## 2. Backend — Fix Upload Validation

- [ ] 2.1 Replace `src.Read(fileData)` with `io.ReadAll(src)` in `UploadTemplate` (template.go)
- [ ] 2.2 Set `tpl.Validated` based on validation result in `UploadTemplate`

## 3. Backend — Re-validation on Mapping Update

- [ ] 3.1 Add re-validation logic to `UpdateTemplateMapping` in contract.go: after saving, if `tpl.FilePath != ""`, read file and validate against new activeFields, update `tpl.Validated`

## 4. Backend — Export Gate

- [ ] 4.1 Add `Validated` check in `ExportContract`: if `tpl.FilePath != "" && !tpl.Validated`, return 409 with error message

## 5. Frontend — Template Type Update

- [ ] 5.1 Add `validated` field to `Template` interface in `frontend/src/api/index.ts`

## 6. Frontend — Validation Status in Settings

- [ ] 6.1 Display validation status indicator in Settings.vue template list (validated = true/false)
- [ ] 6.2 After mapping update, show validation result notification (success or warning with missing fields)

## 7. Frontend — Export Error Handling

- [ ] 7.1 Handle 409 error in contract export flow with user-friendly notification about template validation

## 8. Verification

- [ ] 8.1 Run `go build ./...` to verify backend compiles
- [ ] 8.2 Run `vue-tsc --noEmit` and `npm run build` to verify frontend compiles
