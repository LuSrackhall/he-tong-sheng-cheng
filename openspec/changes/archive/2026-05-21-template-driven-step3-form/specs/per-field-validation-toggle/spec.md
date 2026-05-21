## ADDED Requirements

### Requirement: Per-field validation toggle in active fields summary

The system SHALL provide a per-field validation toggle on each field chip in the "已启用字段" (Enabled Fields) summary area, allowing users to independently control whether a field participates in Word placeholder validation.

#### Scenario: Toggle validation off for a field
- **WHEN** the user clicks the `✓校验` indicator on an enabled field chip in the summary area
- **THEN** the indicator changes to `✗不校验`, `activeFields[fieldKey]` is set to `false`, and the field is excluded from Word validation while still participating in rendering

#### Scenario: Toggle validation on for a field
- **WHEN** the user clicks the `✗不校验` indicator on an enabled field chip
- **THEN** the indicator changes to `✓校验`, `activeFields[fieldKey]` is set to `true`, and the field participates in Word validation again

#### Scenario: Validation toggle only visible for enabled fields
- **WHEN** a field is not enabled (not in activeFields)
- **THEN** the field does not appear in the summary area and no validation toggle is shown

### Requirement: JSON editor and chip bidirectional sync for validation toggle

The system SHALL keep the validation toggle state consistent between the chip UI and the activeFields JSON data. Changes from either side SHALL be reflected in the other.

#### Scenario: Chip toggle updates activeFields
- **WHEN** the user toggles validation off via chip for field `yearlyRent`
- **THEN** `activeFields["yearlyRent"]` becomes `false`, and on next save the updated activeFields is persisted to the backend

#### Scenario: Manual JSON edit affects chip display
- **WHEN** the user manually edits the fieldMap JSON and saves, causing activeFields to be re-parsed
- **THEN** the summary chips reflect the new validation states from the updated activeFields

### Requirement: Validation only checks fields with validate=true

The system SHALL only validate Word document placeholders against fields where `activeFields[key] === true`. Fields with `activeFields[key] === false` SHALL be skipped during validation.

#### Scenario: Field with validate=false is skipped in upload validation
- **WHEN** a Word document is uploaded and `yearlyRent` has `activeFields["yearlyRent"] = false`
- **THEN** the system does not check for `${yearlyRent}` placeholder in the document, and missing it does not cause validation failure

#### Scenario: Field with validate=true must be present
- **WHEN** a Word document is uploaded and `monthlyRent` has `activeFields["monthlyRent"] = true`
- **THEN** the system checks for `${monthlyRent}` placeholder and returns 400 with missingFields if absent

#### Scenario: Re-validation on mapping update respects validation toggle
- **WHEN** the user updates field mapping and a field's validation is toggled off
- **THEN** the re-validation skips that field, and the template's `validated` status is determined only by fields with validate=true
