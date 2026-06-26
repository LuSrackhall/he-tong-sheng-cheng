## MODIFIED Requirements

### Requirement: Step 3 form fields driven by template activeFields

The system SHALL render Step 3 form fields exclusively based on the template's `activeFields` configuration. Only fields present in `activeFields` SHALL appear in the form. All error messages returned to the frontend SHALL be in Chinese.

#### Scenario: Form renders only active fields
- **WHEN** the user reaches Step 3 and the selected template has `activeFields` = `{"startDate": true, "endDate": true, "monthlyRent": true, "assetName": true, "tenantName": true}`
- **THEN** the form displays exactly those 5 fields and no others

#### Scenario: Field not in activeFields is hidden
- **WHEN** the template's activeFields does not include `deposit`
- **THEN** the deposit input is not rendered in Step 3, even though its value still defaults to 0 in the backend

#### Scenario: No template selected
- **WHEN** the user reaches Step 3 without selecting a template
- **THEN** the form falls back to the minimum required fields: startDate, endDate, monthlyRent

#### Scenario: Error messages are in Chinese
- **WHEN** the backend returns an error response (e.g., missing required fields, template validation failure)
- **THEN** the error message SHALL be in Chinese (e.g., "缺少必填字段映射" not "Missing required field mapping")
