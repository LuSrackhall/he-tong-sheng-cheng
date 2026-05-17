# System Printing

## ADDED Requirements

### Requirement: System SHALL generate contract PDF by replacing template placeholders

The system MUST load the user-uploaded .docx template, replace all mapped placeholders with actual data from Contract, Asset, and Tenant entities, and render as PDF.

#### Scenario: Generate contract PDF with full data
- **WHEN** user requests contract PDF for contract C001 with associated asset and tenant
- **THEN** the system returns a PDF where all placeholders (e.g., {{tenantName}}, {{assetName}}, {{startDate}}) are replaced with actual values

#### Scenario: Template has unmapped placeholders
- **WHEN** a template contains a placeholder not mapped to any system field
- **THEN** the placeholder is left as-is in the output with a visual warning marker

### Requirement: System SHALL generate ID card copy PDF at original dimensions

The system MUST render the tenant's ID card image at 85.6mm × 54mm on A4 paper with a watermark annotation "仅用于租赁合同备案".

#### Scenario: Generate ID card copy
- **WHEN** user requests ID card copy for a tenant with an uploaded ID card image
- **THEN** a PDF is returned with the image scaled to 85.6mm × 54mm centered on A4, annotated with "仅用于租赁合同备案"

#### Scenario: Tenant has no ID card image
- **WHEN** user requests ID card copy for a tenant without an uploaded ID card image
- **THEN** the system returns an error indicating the image is missing

### Requirement: System SHALL manage receipt books with auto-incrementing sequence numbers

Each ReceiptBook has a prefix, start number, current number, and total pages. The system MUST auto-increment CurrentNum on each receipt print.

#### Scenario: Create a new receipt book
- **WHEN** admin creates a receipt book with prefix "NO.", start number 1, total pages 50
- **THEN** the book is created with CurrentNum=1 and status "使用中"

#### Scenario: Print receipt auto-increments sequence number
- **WHEN** a receipt is printed using a book with CurrentNum=5
- **THEN** the receipt shows sequence number 5 and the book's CurrentNum becomes 6

#### Scenario: Receipt book exhausted
- **WHEN** a receipt is printed and CurrentNum exceeds the book's total pages
- **THEN** the book status becomes "已用完" and the system prevents further printing with this book

### Requirement: System SHALL generate triple-copy receipt PDF on A4 portrait layout

The receipt PDF MUST be A4 portrait divided into three equal sections (approx 99mm × 210mm each), labeled "存根联" / "收据联" / "记账联", with cut lines between sections.

#### Scenario: Generate triple-copy receipt
- **WHEN** user prints a receipt for a payment of 3000 for contract C001
- **THEN** the returned A4 PDF has three equal sections, each containing the same payment data but labeled "存根联", "收据联", "记账联" respectively, with dashed cut lines between sections

#### Scenario: Receipt content correctness
- **WHEN** a receipt is printed for payment of 5000 for contract C001 with tenant 张三
- **THEN** each section of the receipt displays: receipt book prefix + sequence number, tenant name 张三, contract ID C001, amount 5000, payment date, and the corresponding section label
