# User Auth

## ADDED Requirements

### Requirement: System SHALL support JWT-based authentication

Users MUST authenticate with username and password to receive a JWT token. All API endpoints except login SHALL require a valid token.

#### Scenario: Successful login
- **WHEN** a user logs in with valid credentials
- **THEN** the system returns a JWT token with user ID and role embedded

#### Scenario: Failed login
- **WHEN** a user logs in with invalid credentials
- **THEN** the system returns 401 with error message

#### Scenario: Access protected endpoint without token
- **WHEN** a request is made to any non-login API endpoint without a JWT token
- **THEN** the system returns 401

#### Scenario: Access protected endpoint with expired token
- **WHEN** a request is made with an expired JWT token
- **THEN** the system returns 401

### Requirement: System SHALL support role-based authorization with admin and operator roles

The system MUST distinguish between admin and operator roles. Admin can manage users; operator can perform daily operations (create contracts, record payments, view arrears lists).

#### Scenario: Admin creates new user
- **WHEN** an admin user creates a new user account
- **THEN** the user is created with the specified role

#### Scenario: Operator cannot manage users
- **WHEN** an operator attempts to access user management endpoints
- **THEN** the system returns 403

#### Scenario: Admin can delete a user
- **WHEN** an admin deletes an existing user
- **THEN** the user account is removed

### Requirement: System SHALL include a default admin account on first run

On first startup with an empty database, the system MUST create a default admin account.

#### Scenario: First run creates default admin
- **WHEN** the system starts for the first time with an empty database
- **THEN** a default admin account is created with credentials displayed in the startup log

#### Scenario: Subsequent startups do not create duplicate admin
- **WHEN** the system starts again with an existing admin account
- **THEN** no new default account is created

### Requirement: System SHALL provide a login endpoint to retrieve current user info

An authenticated user MUST be able to retrieve their own user information.

#### Scenario: Get current user info
- **WHEN** an authenticated user requests their own info
- **THEN** the system returns the user's ID, username, and role
