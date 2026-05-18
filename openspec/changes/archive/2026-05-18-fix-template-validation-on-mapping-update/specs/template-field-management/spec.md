## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: All Active Fields Must Be Validated on Upload

The system SHALL validate that ALL fields in `activeFields` have corresponding placeholders in the uploaded Word document, using complete file reading. A field that is disabled (not in `activeFields`) SHALL NOT participate in replacement or validation.

#### Scenario: Upload passes with all active fields present
- **WHEN** a Word document is uploaded and all active fields have matching `${fieldName}` placeholders in the document
- **THEN** the upload succeeds, the file is saved, and the template's `validated` status is set to `true`

#### Scenario: Upload fails when any active field is missing
- **WHEN** a Word document is uploaded and one or more active fields are missing from the document
- **THEN** the system returns 400 with a list of missing field names and the template's `validated` status is set to `false`

#### Scenario: Disabled field missing is allowed
- **WHEN** a Word document is uploaded, a field is disabled (not in `activeFields`), and that field's placeholder is missing from the document
- **THEN** the upload succeeds because disabled fields are not validated
