## MODIFIED Requirements

### Requirement: Step 3 form fields driven by template fieldMap

The system SHALL render Step 3 form fields based on the template's `fieldMap` configuration (JSON mapping of field names to Chinese labels). All fields in `fieldMap` SHALL appear in the form regardless of their `activeFields` validation status.

#### Scenario: Form renders all fieldMap fields
- **WHEN** the user reaches Step 3 and the selected template has `fieldMap` = `{"startDate": "开始日期", "endDate": "结束日期", "monthlyRent": "月租金", "备注": "备注"}`
- **THEN** the form displays all 4 fields, even if `activeFields` has `{"startDate": true, "endDate": true, "monthlyRent": true, "备注": false}`

#### Scenario: Field not in activeFields still renders
- **WHEN** a field exists in `fieldMap` but its `activeFields` value is `false`
- **THEN** the field still renders in the form as an editable input

#### Scenario: Field labels from fieldMap take priority
- **WHEN** `fieldMap` contains `{"monthlyRent": "月租金（含税）"}`
- **THEN** the form displays "月租金（含税）" as the label, not the hardcoded "月租金"

#### Scenario: Fallback when fieldMap is empty
- **WHEN** the template has no `fieldMap` (or fieldMap is empty/null)
- **THEN** the form falls back to using `activeFields` keys as the field list, with hardcoded labels

#### Scenario: Fallback when fieldMap parse fails
- **WHEN** `fieldMap` is not valid JSON
- **THEN** the form falls back to using `activeFields` keys, with hardcoded labels

#### Scenario: No template selected
- **WHEN** the user reaches Step 3 without selecting a template
- **THEN** the form falls back to the minimum required fields: startDate, endDate, monthlyRent

### Requirement: Form fields classified by value source

The system SHALL render each form field as read-only or editable based on its value source category.

#### Scenario: System-auto fields are read-only
- **WHEN** a field like `contractId`, `totalReceivable`, `totalReceived`, `status`, or `today` is in fieldMap
- **THEN** the field renders as a disabled input with gray background, showing the auto-generated or computed value

#### Scenario: Asset/tenant fields are read-only
- **WHEN** a field like `assetName`, `assetType`, `assetDescription`, `tenantName`, `tenantIDCard`, or `tenantPhone` is in fieldMap
- **THEN** the field renders as a disabled input showing the value from the selected asset or tenant

#### Scenario: User-input fields are editable
- **WHEN** a field like `startDate`, `endDate`, `monthlyRent`, `yearlyRent`, `deposit`, or `notes` is in fieldMap
- **THEN** the field renders as an editable input

#### Scenario: Custom fields are editable
- **WHEN** a custom field (not in the preset field list) is in fieldMap
- **THEN** the field renders as an editable text input with the label from fieldMap
