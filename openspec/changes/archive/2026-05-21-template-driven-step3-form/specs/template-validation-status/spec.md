## MODIFIED Requirements

### Requirement: Validation on Word Upload

The system SHALL validate fields with `activeFields[key] === true` against Word document placeholders when a .docx file is uploaded, and persist the validation result to the Template record. Fields with `activeFields[key] === false` SHALL be skipped during validation.

#### Scenario: Upload passes validation
- **WHEN** a .docx file is uploaded and all fields with `activeFields[key] === true` have matching `${fieldName}` placeholders in the document
- **THEN** the file is saved and `validated` is set to `true` on the template

#### Scenario: Upload fails validation
- **WHEN** a .docx file is uploaded and one or more fields with `activeFields[key] === true` are missing from the document
- **THEN** the system returns 400 with a list of missing field names and `validated` is set to `false`

#### Scenario: Upload with no validate=true fields
- **WHEN** a .docx file is uploaded and no fields have `activeFields[key] === true` (all are false or activeFields is empty)
- **THEN** the file is saved without placeholder validation and `validated` is set to `true`

#### Scenario: Upload with validate=false fields only
- **WHEN** a .docx file is uploaded, `yearlyRent` has `activeFields["yearlyRent"] = false`, and `${yearlyRent}` is missing from the document
- **THEN** validation succeeds because `yearlyRent` is excluded from validation

### Requirement: Re-validation on Mapping Update

The system SHALL re-validate the existing Word file when field mapping or active fields are updated, validating only fields with `activeFields[key] === true`, and update the `validated` status accordingly.

#### Scenario: Mapping update triggers re-validation
- **WHEN** the user updates `fieldMap` or `activeFields` and a Word file has already been uploaded for the template
- **THEN** the system re-validates the existing file against only fields where `activeFields[key] === true` and updates `validated` on the template

#### Scenario: Mapping update with no uploaded file
- **WHEN** the user updates `fieldMap` or `activeFields` but no Word file has been uploaded yet
- **THEN** the system saves the mapping without validation and `validated` remains `false`

#### Scenario: Mapping update with missing file on disk
- **WHEN** the user updates mapping, the `filePath` is set but the actual file does not exist on disk
- **THEN** the system saves the mapping without validation and `validated` remains `false`

### Requirement: Export Blocked for Unvalidated Templates

The system SHALL block contract export when the template's validation status is `false`, and provide a user-friendly error message. The system SHALL also block export when required fields (`startDate`, `endDate`, `monthlyRent`, `tenantName`, `assetName`) are not present in `activeFields`.

#### Scenario: Export blocked when template is not validated
- **WHEN** the user attempts to export a contract whose template has `validated = false`
- **THEN** the system returns 409 with an error message indicating the template validation has not passed and the user should upload a compliant Word file

#### Scenario: Export allowed when template is validated
- **WHEN** the user attempts to export a contract whose template has `validated = true` and all required fields are in activeFields
- **THEN** the export proceeds normally

#### Scenario: Export with no template file
- **WHEN** the user attempts to export a contract whose template has no uploaded file (`filePath` is empty)
- **THEN** the system returns 400 with "Template file not uploaded yet"

#### Scenario: Export blocked when required fields missing from activeFields
- **WHEN** the user attempts to export a contract whose template is missing `assetName` from activeFields
- **THEN** the system returns 409 with an error indicating the required field is missing
