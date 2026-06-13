## ADDED Requirements

### Requirement: Gitignore shall exclude common sensitive patterns

The root `.gitignore` SHALL include rules for `.env` / `.env.*`, `*.log`, `.DS_Store`, `__pycache__/`, `*.pyc`, `tmp/`, and `node_modules/`.

#### Scenario: Developer creates .env file
- **WHEN** a developer creates `.env` or `.env.local` in the project root
- **THEN** git SHALL NOT track the file and `git status` SHALL NOT show it as untracked

#### Scenario: macOS creates .DS_Store
- **WHEN** macOS generates `.DS_Store` files in any directory
- **THEN** git SHALL NOT track them

#### Scenario: Python generates cache files
- **WHEN** Python creates `__pycache__/` directories or `*.pyc` files
- **THEN** git SHALL NOT track them

#### Scenario: Log files are generated
- **WHEN** the application or tools generate `*.log` files
- **THEN** git SHALL NOT track them

### Requirement: Build artifacts shall not be tracked

The `cmd/server/dist/` directory SHALL NOT be tracked by git. Existing tracked files in that path SHALL be removed from the git index.

#### Scenario: dist files removed from index
- **WHEN** `git rm --cached -r cmd/server/dist/` is executed and committed
- **THEN** `git ls-files cmd/server/dist/` SHALL return empty output
- **AND** the local `cmd/server/dist/` directory SHALL remain intact on disk

#### Scenario: Frontend build does not re-add dist
- **WHEN** `npm run build` is run (which outputs to `cmd/server/dist/`)
- **THEN** `git status` SHALL NOT show any files under `cmd/server/dist/` as untracked or modified

### Requirement: JWT secret shall require explicit configuration

The application SHALL require the `JWT_SECRET` environment variable to be set. If unset or empty, the application SHALL log a fatal error and exit.

#### Scenario: JWT_SECRET is set
- **WHEN** the server starts with `JWT_SECRET=my-production-secret`
- **THEN** the server SHALL start normally and use that value as the JWT signing key

#### Scenario: JWT_SECRET is unset
- **WHEN** the server starts without `JWT_SECRET` environment variable
- **THEN** the server SHALL log `FATAL: JWT_SECRET environment variable is required` and exit with code 1

#### Scenario: JWT_SECRET is empty string
- **WHEN** the server starts with `JWT_SECRET=""`
- **THEN** the server SHALL log `FATAL: JWT_SECRET environment variable is required` and exit with code 1
