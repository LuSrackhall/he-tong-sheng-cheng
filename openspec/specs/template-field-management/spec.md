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

The system SHALL automatically mark a field as enabled when it is added to the field map, without requiring a separate toggle action.

#### Scenario: Preset field label click auto-enables
- **WHEN** the user clicks a preset field label to add it to the field map
- **THEN** the field is immediately added to both `fieldMap` and `activeFields` (activeSet), so it participates in both replacement and validation

#### Scenario: Custom field auto-enables on creation
- **WHEN** the user creates a new custom field
- **THEN** the field is automatically added to `activeFields` and its toggle is shown as enabled

### Requirement: All Active Fields Must Be Validated on Upload

The system SHALL validate that ALL fields in `activeFields` have corresponding placeholders in the uploaded Word document. A field that is disabled (not in `activeFields`) SHALL NOT participate in replacement or validation.

#### Scenario: Upload passes with all active fields present
- **WHEN** a Word document is uploaded and all active fields have matching `${fieldName}` placeholders in the document
- **THEN** the upload succeeds and the file is saved

#### Scenario: Upload fails when any active field is missing
- **WHEN** a Word document is uploaded and one or more active fields are missing from the document
- **THEN** the system returns 400 with a list of missing field names

#### Scenario: Disabled field missing is allowed
- **WHEN** a Word document is uploaded, a field is disabled (not in `activeFields`), and that field's placeholder is missing from the document
- **THEN** the upload succeeds because disabled fields are not validated

### Requirement: Field Label Display

The system SHALL display both the placeholder name and its human-readable label in the template field mapping UI.

#### Scenario: Field labels are shown alongside placeholders
- **WHEN** the user views the field mapping section of a template
- **THEN** each field is displayed as `${fieldName} → 显示标签` (e.g., `${tenantName} → 租户姓名`), using the label from `fieldMap`

#### Scenario: Fields without labels fall back to field name
- **WHEN** a field in `fieldMap` has no display label (empty string value)
- **THEN** the UI displays only `${fieldName}` without the arrow or label

