# Tenant Management

## ADDED Requirements

### Requirement: System SHALL support creating tenants with ID card information

A tenant can be created with name, ID number, gender, phone, address. All fields except name and ID number are optional.

#### Scenario: Create tenant with full information
- **WHEN** user creates a tenant with name, ID number, gender, phone, and address
- **THEN** the tenant is created successfully and all fields are stored

#### Scenario: Create tenant with minimal information
- **WHEN** user creates a tenant with only name and ID number
- **THEN** the tenant is created successfully with gender defaulting to "未知"

### Requirement: System SHALL support tenant listing and search

Users can list all tenants and search by name, ID number, or phone.

#### Scenario: Search tenants by name
- **WHEN** user searches tenants with keyword "张三"
- **THEN** all tenants whose name contains "张三" are returned

#### Scenario: Search tenants by ID number
- **WHEN** user searches tenants with partial ID number "3201"
- **THEN** all tenants whose ID number contains "3201" are returned

### Requirement: System SHALL support ExtraFields for extensible tenant attributes

Tenant non-core fields are stored as JSON in ExtraFields. The field list is configured in system settings. System-suggested fields (name, ID number, address, phone, gender) are shown via UI prompts for user consideration.

#### Scenario: Display tenant with configured extra fields
- **WHEN** system settings define extra fields ["职业", "紧急联系人"] for tenants
- **THEN** the tenant detail view renders input controls for each configured field alongside core fields

#### Scenario: System-suggested fields are displayed as hints
- **WHEN** user sets up tenant fields
- **THEN** the UI shows suggested fields (姓名, 身份证号, 住址, 手机号, 称谓) as prompts the user may choose to include

### Requirement: System SHALL track historical lease records per tenant

Each tenant's detail page MUST display all contracts historically associated with them.

#### Scenario: View tenant leasing history
- **WHEN** user views a tenant who has signed 3 contracts
- **THEN** all 3 contracts are listed with their status, asset name, and date range
