## ADDED Requirements

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
