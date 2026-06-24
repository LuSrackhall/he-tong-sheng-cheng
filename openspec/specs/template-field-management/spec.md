# template-field-management Specification

## Purpose
TBD - created by archiving change improve-template-management. Update Purpose after archive.
## Requirements
### Requirement: Custom Field Addition

The system SHALL allow users to add custom mapping fields with a field name and display label via the template settings UI.

#### Scenario: User adds a custom field
- **WHEN** the user clicks the "添加自定义字段" button, enters a unique field name and display label, and confirms
- **THEN** the field is added to the `fieldMap` JSON as `{"fieldName": "displayLabel"}`, appended to the preset field list in the UI, and the field is automatically added to `activeFields`

#### Scenario: Duplicate field name is rejected
- **WHEN** the user attempts to add a custom field whose name already exists in the preset or custom fields
- **THEN** the system displays a validation error indicating the field name already exists, and the field is not added

#### Scenario: Empty field name or label is rejected
- **WHEN** the user submits the custom field form with an empty field name or display label
- **THEN** the system displays a validation error and the field is not added

### Requirement: Auto-Enable on Field Addition

The system SHALL automatically mark a field as enabled when it is added to the field map, without requiring a separate toggle action. Newly added fields SHALL default to `activeFields[key] = true` (enabled and validated).

#### Scenario: Preset field label click auto-enables
- **WHEN** the user clicks a preset field label to add it to the field map
- **THEN** the field is immediately added to both `fieldMap` and `activeFields` with value `true`, so it participates in both replacement and validation

#### Scenario: Custom field auto-enables on creation
- **WHEN** the user creates a new custom field
- **THEN** the field is automatically added to `activeFields` with value `true`

### Requirement: All Active Fields Must Be Validated on Upload

The system SHALL validate that fields with `activeFields[key] === true` have corresponding placeholders in the uploaded Word document. A field that is disabled (not in `activeFields`) SHALL NOT participate in replacement or validation. A field with `activeFields[key] === false` SHALL participate in replacement but NOT in validation.

#### Scenario: Upload passes with all validate=true fields present
- **WHEN** a Word document is uploaded and all fields with `activeFields[key] === true` have matching `${fieldName}` placeholders in the document
- **THEN** the upload succeeds, the file is saved, and the template's `validated` status is set to `true`

#### Scenario: Upload fails when any validate=true field is missing
- **WHEN** a Word document is uploaded and one or more fields with `activeFields[key] === true` are missing from the document
- **THEN** the system returns 400 with a list of missing field names and the template's `validated` status is set to `false`

#### Scenario: Validate=false field missing is allowed
- **WHEN** a Word document is uploaded, a field has `activeFields[key] === false`, and that field's placeholder is missing from the document
- **THEN** the upload succeeds because only validate=true fields are validated

#### Scenario: Disabled field missing is allowed
- **WHEN** a Word document is uploaded, a field is not in `activeFields`, and that field's placeholder is missing from the document
- **THEN** the upload succeeds because disabled fields are not validated

### Requirement: Field Label Display

The system SHALL display both the placeholder name and its human-readable label in the template field mapping UI.

#### Scenario: Field labels are shown alongside placeholders
- **WHEN** the user views the field mapping section of a template
- **THEN** each field is displayed as `${fieldName} → 显示标签` (e.g., `${tenantName} → 租户姓名`), using the label from `fieldMap`

#### Scenario: Fields without labels fall back to field name
- **WHEN** a field in `fieldMap` has no display label (empty string value)
- **THEN** the UI displays only `${fieldName}` without the arrow or label

### Requirement: Re-validation on Mapping Update

The system SHALL re-validate the existing Word file against the new active fields whenever field mapping or active fields are updated, and update the template's validation status accordingly.

#### Scenario: Mapping update with existing file passes validation
- **WHEN** the user updates `fieldMap` or `activeFields`, a Word file has already been uploaded, and all new active fields are present in the document
- **THEN** the mapping is saved and the template's `validated` status is set to `true`

#### Scenario: Mapping update with existing file fails validation
- **WHEN** the user updates `fieldMap` or `activeFields`, a Word file has already been uploaded, and one or more new active fields are missing from the document
- **THEN** the mapping is saved successfully (not blocked), but the template's `validated` status is set to `false`

#### Scenario: Mapping update with no uploaded file
- **WHEN** the user updates `fieldMap` or `activeFields` and no Word file has been uploaded yet
- **THEN** the mapping is saved and `validated` remains `false` without attempting validation

---

### Requirement: activeFields data format

The system SHALL store `activeFields` as a JSON object mapping field keys to boolean validation flags (`Record<string, boolean>`). The system SHALL support backward compatibility by automatically converting legacy `string[]` format to the object format, defaulting all values to `true`.

#### Scenario: activeFields stored as object
- **WHEN** the user saves field mapping
- **THEN** `activeFields` is persisted as `{"startDate": true, "monthlyRent": true, "yearlyRent": false}`

#### Scenario: Legacy array format auto-converted
- **WHEN** the system reads a template whose `activeFields` is `["startDate", "monthlyRent"]`
- **THEN** it is treated as `{"startDate": true, "monthlyRent": true}`

### Requirement: activeFields sync with fieldMap on save

The system SHALL automatically synchronize `activeFields` with `fieldMap` when saving mapping: fields present in uncommented fieldMap SHALL be added to activeFields with default `true`, and fields removed from fieldMap SHALL be removed from activeFields.

#### Scenario: New field auto-added to activeFields
- **WHEN** the user adds a new field to fieldMap and saves
- **THEN** the field appears in activeFields with value `true`

#### Scenario: Removed field auto-removed from activeFields
- **WHEN** the user removes a field from fieldMap (via chip click or JSON edit) and saves
- **THEN** the field is removed from activeFields

#### Scenario: Existing validation toggle preserved on save
- **WHEN** the user has set `yearlyRent` validation to `false`, then adds a new field and saves
- **THEN** `activeFields["yearlyRent"]` remains `false` while the new field gets `true`

### Requirement: yearlyRent in preset field groups

The system SHALL include `yearlyRent` in the "合同类" preset field group alongside `monthlyRent`, with display label "年租金".

#### Scenario: yearlyRent visible in preset fields
- **WHEN** the user views the template field mapping section
- **THEN** `yearlyRent` appears as a preset chip in the "合同类" group with label "年租金"

#### Scenario: yearlyRent can be added to mapping
- **WHEN** the user clicks the `yearlyRent` preset chip
- **THEN** it is added to fieldMap as `"yearlyRent": "年租金"` and to activeFields as `"yearlyRent": true`

### Requirement: signingDate in preset field groups

The system SHALL include `signingDate` in the "合同类" preset field group, with display label "签订日期".

#### Scenario: signingDate visible in preset fields
- **WHEN** the user views the template field mapping section
- **THEN** `signingDate` appears as a preset chip in the "合同类" group with label "签订日期"

#### Scenario: signingDate can be added to mapping
- **WHEN** the user clicks the `signingDate` preset chip
- **THEN** it is added to fieldMap as `"signingDate": "签订日期"` and to activeFields as `"signingDate": true`

### Requirement: signingDate template replacement value

The system SHALL replace `${signingDate}` placeholders in Word templates with the contract's creation date in `YYYY-MM-DD` format.

#### Scenario: signingDate replaces with contract creation date
- **WHEN** a Word template containing `${signingDate}` is rendered for a contract
- **THEN** the placeholder is replaced with the contract's `CreatedAt` formatted as `2006-01-02`

#### Scenario: signingDate available in template preview
- **WHEN** the user previews a template
- **THEN** the field list includes `signingDate` with label "签订日期"

### Requirement: signingDate in builtin field keys

The system SHALL recognize `signingDate` as a builtin field key, preventing it from being treated as a custom field in the template preview.

#### Scenario: signingDate not duplicated as custom field
- **WHEN** the template has `signingDate` in its fieldMap
- **THEN** the template preview does not list it as a custom field (it appears only in the builtin fields section)

