# Contract Management

## ADDED Requirements

### Requirement: System SHALL calculate total receivable amount from contract dates and monthly rent

The system MUST automatically calculate TotalReceivable from startDate, endDate, and monthlyRent. The result can be manually adjusted to accommodate offline negotiations like first-month half-price.

#### Scenario: Full month calculation
- **WHEN** a contract starts on 2026-01-01 and ends on 2026-12-31 with monthlyRent 1000
- **THEN** TotalReceivable is calculated as 12 × 1000 = 12000

#### Scenario: Partial month at end
- **WHEN** a contract starts on 2026-01-01 and ends on 2026-02-15 with monthlyRent 1000
- **THEN** TotalReceivable is calculated as 1 × 1000 + 15 × (1000/30) = 1500

#### Scenario: Manual adjustment of total receivable
- **WHEN** user manually sets TotalReceivable to 11000 for a contract with calculated amount 12000
- **THEN** TotalReceivable is saved as 11000 and used for all subsequent calculations

### Requirement: System SHALL support contract template upload with field mapping

Users upload .docx templates and define mappings from template placeholders to system fields.

#### Scenario: Upload and configure a contract template
- **WHEN** user uploads a .docx file and defines mapping {`{{tenantName}}`: "tenant.name", `{{assetName}}`: "asset.name"}
- **THEN** the template is stored and field mapping is saved as JSON

#### Scenario: Generate contract PDF from template
- **WHEN** user triggers contract printing for a specific contract
- **THEN** the system replaces all mapped placeholders with actual data and returns a PDF

### Requirement: System SHALL manage contract lifecycle with status transitions

Contract statuses SHALL be: 执行中 (active), 已缴全 (fully paid), 欠费中 (in arrears), 已到期 (expired).

#### Scenario: Contract starts as active
- **WHEN** a new contract is created with future end date
- **THEN** its status is "执行中"

#### Scenario: Contract becomes fully paid
- **WHEN** totalReceived becomes equal to or greater than totalReceivable before endDate
- **THEN** its status automatically becomes "已缴全"

#### Scenario: Contract becomes expired with arrears
- **WHEN** endDate passes and totalReceived < totalReceivable
- **THEN** its status automatically becomes "已到期"

### Requirement: System SHALL associate every payment record with a contract

Each contract detail view MUST show all associated payment records.

#### Scenario: View contract payment history
- **WHEN** user views a contract that has received 5 payments
- **THEN** all 5 payment records are displayed with amount and date
