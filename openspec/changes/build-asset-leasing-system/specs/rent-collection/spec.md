# Rent Collection

## ADDED Requirements

### Requirement: System SHALL record each payment against a specific contract

Each payment is independently recorded with amount and timestamp, associated to a contract. The system does NOT assume any payment frequency (monthly/quarterly/yearly).

#### Scenario: Record a payment for an active contract
- **WHEN** user records a payment of 3000 for contract C001
- **THEN** a Payment record is created with ContractID=C001, Amount=3000, PaidAt=current timestamp

#### Scenario: Payment amount is independent of monthly rent
- **WHEN** user records a payment of 5000 for a contract with monthlyRent 1000
- **THEN** the payment is accepted and recorded as 5000 regardless of the monthly rent amount

### Requirement: System SHALL update contract total received on each payment

After each payment is recorded, the contract's TotalReceived is updated by accumulating the payment amount.

#### Scenario: First payment updates total received
- **WHEN** a first payment of 3000 is recorded for a contract with TotalReceived=0
- **THEN** the contract's TotalReceived becomes 3000

#### Scenario: Multiple payments accumulate
- **WHEN** three payments of 3000 each are recorded for the same contract
- **THEN** the contract's TotalReceived becomes 9000

### Requirement: System SHALL recalculate "money used up date" after each payment

The system MUST calculate the date through which the tenant has paid by converting total received into whole calendar months plus remainder days.

#### Scenario: Exact months paid
- **WHEN** totalReceived is 3000 and monthlyRent is 1000, starting from 2026-01-01
- **THEN** the money used up date is 2026-04-01 (3 full calendar months from start)

#### Scenario: Months plus partial days
- **WHEN** totalReceived is 3500 and monthlyRent is 1000, starting from 2026-01-01
- **THEN** the money used up date is 2026-04-16 (3 months + ceil(500 / (1000/30)) = 15 days from April 1)

#### Scenario: Used up date capped at end date
- **WHEN** totalReceived would push the used up date beyond the contract end date
- **THEN** the used up date is capped at the contract end date

### Requirement: System SHALL mark contract as fully paid when total received reaches total receivable

When TotalReceived >= TotalReceivable before endDate, the contract status MUST automatically become "已缴全".

#### Scenario: Contract fully paid before end date
- **WHEN** totalReceived reaches 12000 for a contract with TotalReceivable=12000 and endDate in the future
- **THEN** the contract status becomes "已缴全"

### Requirement: System SHALL display arrears gap during payment recording

When recording a payment, the system MUST show how much is still owed (TotalReceivable - TotalReceived).

#### Scenario: Show arrears gap after payment
- **WHEN** user records a payment for a contract with TotalReceivable=12000 and TotalReceived=6000
- **THEN** the UI displays "还差 6000" (still short 6000)

### Requirement: System SHALL support receipt printing after payment

After recording a payment, user can print a triple-copy receipt with auto-incrementing sequence number managed by ReceiptBook.

#### Scenario: Print receipt after payment
- **WHEN** user triggers receipt printing after recording a payment
- **THEN** a PDF receipt is generated with the current ReceiptBook's next sequence number, and the ReceiptBook CurrentNum is incremented