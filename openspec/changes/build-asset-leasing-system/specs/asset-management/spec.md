# Asset Management

## ADDED Requirements

### Requirement: System SHALL support creating assets with minimal fields

An asset can be created with only a name and monthly rent. All other fields are optional and can be filled in later.

#### Scenario: Create asset with minimal fields
- **WHEN** user creates a new asset with name and monthlyRent only
- **THEN** the asset is created successfully and can be used immediately in contract creation

#### Scenario: Create asset without monthlyRent
- **WHEN** user attempts to create an asset without providing monthlyRent
- **THEN** the system rejects the request with a validation error

### Requirement: System SHALL support asset listing and search

Users can list all assets and search by name, group, or tag.

#### Scenario: Search assets by name
- **WHEN** user searches assets with keyword "商铺"
- **THEN** all assets whose name contains "商铺" are returned

#### Scenario: Filter assets by group
- **WHEN** user filters assets by group "A区"
- **THEN** only assets belonging to group "A区" are returned

### Requirement: System SHALL support ExtraFields for extensible asset attributes

Asset non-core fields are stored as JSON in ExtraFields. The field list is configured in system settings and the UI dynamically renders input controls based on the configuration.

#### Scenario: Display asset with configured extra fields
- **WHEN** system settings define extra fields ["面积", "楼层", "朝向"] for assets
- **THEN** the asset detail view renders input controls for each configured field

#### Scenario: Asset with no extra fields configured
- **WHEN** system settings define no extra fields for assets
- **THEN** the asset detail view shows only name and monthlyRent

### Requirement: System SHALL track historical lease records per asset

Each asset's detail page MUST display all contracts historically associated with it.

#### Scenario: View asset leasing history
- **WHEN** user views an asset that has been associated with 3 contracts
- **THEN** all 3 contracts are listed with their status and date range
