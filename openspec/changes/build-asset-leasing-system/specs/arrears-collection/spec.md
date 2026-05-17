# Arrears Collection

## ADDED Requirements

### Requirement: System SHALL calculate money used up date per contract

The system MUST calculate the date through which the tenant has paid based on total received and monthly rent, using whole calendar months first then remainder days at daily rate (monthlyRent / 30).

#### Scenario: Full calendar months across different month lengths
- **WHEN** totalReceived is 2000, monthlyRent is 1000, start date is 2026-01-31
- **THEN** the used up date is 2026-03-28 (1 full month to 02-28, plus remaining 1000 → 30 days at 33.33/day = 2026-03-28)

#### Scenario: Exact whole months with no remainder
- **WHEN** totalReceived is 3000, monthlyRent is 1000, start date is 2026-01-01
- **THEN** the used up date is 2026-04-01 (3 full calendar months, no remainder)

### Requirement: System SHALL generate five-level arrears classification list daily

The system MUST classify every active contract into one of five levels each day based on "money used up date" and end date. Each contract SHALL appear on exactly one list.

#### Scenario: Level 1 — payment warning
- **WHEN** a contract's used up date is 25 days from now and the contract is not fully paid
- **THEN** the contract appears on the "应缴预警" (Level 1) list

#### Scenario: Level 2 — imminent payment reminder
- **WHEN** a contract's used up date is 5 days from now and the contract is not fully paid
- **THEN** the contract appears on the "近期应缴提醒" (Level 2) list

#### Scenario: Level 3 — overdue collection
- **WHEN** a contract's used up date is 3 days in the past and end date is still in the future
- **THEN** the contract appears on the "逾期未缴催收" (Level 3) list

#### Scenario: Level 4 — expiration warning
- **WHEN** a contract's end date is 20 days from now and totalReceived < totalReceivable
- **THEN** the contract appears on the "到期预警" (Level 4) list

#### Scenario: Level 5 — post-expiration debt recovery
- **WHEN** a contract's end date is 5 days in the past and totalReceived < totalReceivable
- **THEN** the contract appears on the "已到期欠费追缴" (Level 5) list

#### Scenario: Contract appears on highest matching level only
- **WHEN** a contract qualifies for both Level 4 (end date approaching) and Level 3 (overdue)
- **THEN** the contract appears only on Level 3 (the higher priority level)

### Requirement: System SHALL include suggested actions per arrears level

Each arrears list MUST include suggested actions for staff to follow.

#### Scenario: Level 1 suggested action
- **WHEN** viewing the Level 1 "应缴预警" list
- **THEN** each entry shows suggested action "列入观察，心中有数"

#### Scenario: Level 5 suggested action
- **WHEN** viewing the Level 5 "已到期欠费追缴" list
- **THEN** each entry shows suggested action "进入追讨，法律途径"

### Requirement: System SHALL track arrears collection history per contract

Each contract's collection history MUST be traceable, recording which levels it appeared on and when.

#### Scenario: View contract arrears history
- **WHEN** user views arrears history for a contract that escalated from Level 1 to Level 3 over 60 days
- **THEN** the history shows each level change with timestamp
