## ADDED Requirements

### Requirement: Validation on Word Upload

The system SHALL validate all active fields against Word document placeholders when a .docx file is uploaded, and persist the validation result to the Template record.

#### Scenario: Upload passes validation
- **WHEN** a .docx file is uploaded and all active fields have matching `${fieldName}` placeholders in the document
- **THEN** the file is saved and `validated` is set to `true` on the template

#### Scenario: Upload fails validation
- **WHEN** a .docx file is uploaded and one or more active fields are missing from the document
- **THEN** the system returns 400 with a list of missing field names and `validated` is set to `false`

#### Scenario: Upload with no active fields
- **WHEN** a .docx file is uploaded and `activeFields` is empty
- **THEN** the file is saved without placeholder validation and `validated` is set to `true`

### Requirement: Re-validation on Mapping Update

The system SHALL re-validate the existing Word file when field mapping or active fields are updated, and update the `validated` status accordingly.

#### Scenario: Mapping update triggers re-validation
- **WHEN** the user updates `fieldMap` or `activeFields` and a Word file has already been uploaded for the template
- **THEN** the system re-validates the existing file against the new active fields and updates `validated` on the template

#### Scenario: Mapping update with no uploaded file
- **WHEN** the user updates `fieldMap` or `activeFields` but no Word file has been uploaded yet
- **THEN** the system saves the mapping without validation and `validated` remains `false`

#### Scenario: Mapping update with missing file on disk
- **WHEN** the user updates mapping, the `filePath` is set but the actual file does not exist on disk
- **THEN** the system saves the mapping without validation and `validated` remains `false`

### Requirement: Export Blocked for Unvalidated Templates

The system SHALL block contract export when the template's validation status is `false`, and provide a user-friendly error message.

#### Scenario: Export blocked when template is not validated
- **WHEN** the user attempts to export a contract whose template has `validated = false`
- **THEN** the system returns 409 with an error message indicating the template validation has not passed and the user should upload a compliant Word file

#### Scenario: Export allowed when template is validated
- **WHEN** the user attempts to export a contract whose template has `validated = true`
- **THEN** the export proceeds normally

#### Scenario: Export with no template file
- **WHEN** the user attempts to export a contract whose template has no uploaded file (`filePath` is empty)
- **THEN** the system returns 400 with "Template file not uploaded yet" (existing behavior, unchanged)

### Requirement: Validation Status Persistence

The system SHALL persist the `validated` field as a boolean column in the `templates` table, defaulting to `false` for new and existing templates.

#### Scenario: New template has default validated false
- **WHEN** a new template is created
- **THEN** `validated` is `false` by default

#### Scenario: Existing template after migration
- **WHEN** the system runs AutoMigrate after adding the `Validated` field
- **THEN** all existing templates have `validated = false` until they re-upload a Word file
