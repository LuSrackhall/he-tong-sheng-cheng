# yearly-rent-field Specification

## Purpose
TBD - created by archiving change template-driven-step3-form. Update Purpose after archive.
## Requirements
### Requirement: yearlyRent as independent preset field

The system SHALL support `yearlyRent` as an independent preset field in the template field mapping, separate from `monthlyRent`. Users SHALL be able to map `${yearlyRent}` and `${monthlyRent}` independently in their Word templates.

#### Scenario: yearlyRent appears in preset field list
- **WHEN** the user views the field mapping section in template settings
- **THEN** `yearlyRent` appears in the "合同类" preset field group with label "年租金"

#### Scenario: yearlyRent can be enabled independently
- **WHEN** the user enables `yearlyRent` in the template mapping but disables `monthlyRent`
- **THEN** the template's activeFields contains `{"yearlyRent": true}` without `monthlyRent`, and the Word rendering uses only yearlyRent

#### Scenario: Both monthlyRent and yearlyRent enabled
- **WHEN** both fields are enabled in activeFields
- **THEN** the form shows both inputs with linkage enabled by default, and Word rendering outputs both values independently

### Requirement: monthlyRent and yearlyRent linked conversion in Step 3 form

The system SHALL provide automatic linked conversion between monthlyRent and yearlyRent in the Step 3 contract form when both fields are active. The linkage SHALL be enabled by default with a toggle to disconnect.

#### Scenario: Default linked behavior — edit monthlyRent
- **WHEN** both monthlyRent and yearlyRent are active in the form, linkage is on, and the user enters `5000` for monthlyRent
- **THEN** yearlyRent automatically updates to `60000` (monthlyRent × 12)

#### Scenario: Default linked behavior — edit yearlyRent
- **WHEN** both fields are active, linkage is on, and the user enters `60000` for yearlyRent
- **THEN** monthlyRent automatically updates to `5000` (yearlyRent / 12, rounded to 2 decimals)

#### Scenario: Disconnect linkage
- **WHEN** the user clicks the linkage toggle to disconnect
- **THEN** monthlyRent and yearlyRent become independent editable inputs, each keeping its current value

#### Scenario: Reconnect linkage
- **WHEN** the user clicks the linkage toggle to reconnect
- **THEN** yearlyRent is recalculated from the current monthlyRent value (monthlyRent × 12)

#### Scenario: Only one rent field is active
- **WHEN** only monthlyRent or only yearlyRent is active in the form
- **THEN** no linkage toggle is shown, and the single field is simply editable

### Requirement: yearlyRent in Word rendering

The system SHALL populate `${yearlyRent}` in Word rendering with the yearly rent value, calculated as `monthlyRent × 12` if not explicitly set by the user.

#### Scenario: yearlyRent rendered from user input
- **WHEN** the user entered `72000` for yearlyRent (with linkage off)
- **THEN** the Word document renders `${yearlyRent}` as `72000.00`

#### Scenario: yearlyRent rendered from linked calculation
- **WHEN** linkage is on and monthlyRent is `6000`
- **THEN** the Word document renders `${yearlyRent}` as `72000.00`

