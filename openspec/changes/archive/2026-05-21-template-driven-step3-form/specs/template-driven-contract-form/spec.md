## ADDED Requirements

### Requirement: Step 3 form fields driven by template activeFields

The system SHALL render Step 3 form fields exclusively based on the template's `activeFields` configuration. Only fields present in `activeFields` SHALL appear in the form.

#### Scenario: Form renders only active fields
- **WHEN** the user reaches Step 3 and the selected template has `activeFields` = `{"startDate": true, "endDate": true, "monthlyRent": true, "assetName": true, "tenantName": true}`
- **THEN** the form displays exactly those 5 fields and no others

#### Scenario: Field not in activeFields is hidden
- **WHEN** the template's activeFields does not include `deposit`
- **THEN** the deposit input is not rendered in Step 3, even though its value still defaults to 0 in the backend

#### Scenario: No template selected
- **WHEN** the user reaches Step 3 without selecting a template
- **THEN** the form falls back to the minimum required fields: startDate, endDate, monthlyRent

### Requirement: Form fields classified by value source

The system SHALL render each form field as read-only or editable based on its value source category.

#### Scenario: System-auto fields are read-only
- **WHEN** a field like `contractId`, `totalReceivable`, `totalReceived`, `status`, or `today` is active
- **THEN** the field renders as a disabled input with gray background, showing the auto-generated or computed value

#### Scenario: Asset/tenant fields are read-only
- **WHEN** a field like `assetName`, `assetType`, `assetDescription`, `tenantName`, `tenantIDCard`, or `tenantPhone` is active
- **THEN** the field renders as a disabled input showing the value from the selected asset or tenant in Step 1/2

#### Scenario: User-input fields are editable
- **WHEN** a field like `startDate`, `endDate`, `monthlyRent`, `yearlyRent`, `deposit`, or `notes` is active
- **THEN** the field renders as an editable input

#### Scenario: Custom fields are editable
- **WHEN** a custom field (not in the preset field list) is in activeFields
- **THEN** the field renders as an editable text input with the label from fieldMap
