## ADDED Requirements

### Requirement: Template Deletion with Reference Check

The system SHALL allow deletion of a template only when it is not referenced by any existing contract.

#### Scenario: Delete unused template succeeds
- **WHEN** a DELETE request is made to `/api/templates/:id` and the template is not referenced by any contract
- **THEN** the template is physically deleted from the database and a 200 OK response is returned

#### Scenario: Delete referenced template is rejected
- **WHEN** a DELETE request is made to `/api/templates/:id` and at least one contract references this template
- **THEN** the system returns 409 Conflict with an error message indicating the template is in use, and the template is not deleted

#### Scenario: Delete non-existent template
- **WHEN** a DELETE request is made to `/api/templates/:id` for an ID that does not exist
- **THEN** the system returns 404 Not Found

### Requirement: Frontend Delete Button

The frontend MUST provide a delete button for each template in the template management interface.

#### Scenario: User deletes an unused template via UI
- **WHEN** the user clicks the delete button on a template that is not referenced by any contract
- **THEN** a confirmation dialog is shown, and upon confirmation the template is deleted and removed from the list

#### Scenario: User attempts to delete a used template via UI
- **WHEN** the user clicks the delete button on a template that is referenced by at least one contract
- **THEN** an error notification is displayed indicating the template cannot be deleted because it is in use
