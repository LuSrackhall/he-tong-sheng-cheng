# Asset Leasing & Collection Management System — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete contract → payment → arrears collection system with Go backend + Vue 3 frontend, embeddable single binary.

**Architecture:** Go layered architecture (cmd → transport → service → domain ← repository), SQLite/PostgreSQL dual DB via repository interface, Vue 3 SPA embedded via go:embed.

**Tech Stack:** Go + Gin + GORM + JWT, Vue 3 + Vite + Pinia + Vue Router, SQLite/PostgreSQL

---

### Task 1: Go Backend Foundation

**Files:**
- Create: `cmd/server/main.go`
- Create: `internal/config/config.go`
- Create: `internal/domain/asset.go`, `tenant.go`, `contract.go`, `payment.go`, `receipt.go`, `receiptbook.go`, `template.go`, `user.go`
- Create: `internal/domain/repo.go` (repository interfaces)
- Create: `internal/repository/sqlite/setup.go`, `internal/repository/sqlite/*.go`
- Create: `internal/repository/postgres/setup.go`, `internal/repository/postgres/*.go`
- Create: `internal/di/wire.go`

- [ ] 1.1 Initialize Go module, create project skeleton: `cmd/server/main.go`, `internal/` subpackages
- [ ] 1.2 Implement config loading: CLI flags (`--mode`, `--db-host`, `--db-name`), unified config struct
- [ ] 1.3 Define domain entities: Asset, Tenant, Contract, Payment, Receipt, ReceiptBook, Template, User
- [ ] 1.4 Define domain Repository interfaces: all 8 repo interfaces
- [ ] 1.5 Implement SQLite repository: table creation + migration (GORM auto-migrate), all repository implementations
- [ ] 1.6 Implement PostgreSQL repository: all repository implementations
- [ ] 1.7 Implement DI: select SQLite or PostgreSQL per config, wire up dependencies

### Task 2: Core Calculation Engine

**Files:**
- Create: `internal/domain/calc/contract.go`
- Create: `internal/domain/calc/contract_test.go`
- Create: `internal/domain/calc/arrears.go`
- Create: `internal/domain/calc/arrears_test.go`
- Create: `internal/domain/calc/status.go`
- Create: `internal/domain/calc/status_test.go`

- [ ] 2.1 Implement ContractCalc: TotalReceivable, UsedUpDate, Arrears — pure functions + unit tests
- [ ] 2.2 Implement ArrearsCalc: 5-level classification (Level 1-5), arrears list generation — pure functions + unit tests
- [ ] 2.3 Implement ContractStatusCalc: auto status determination (执行中/已缴全/欠费中/已到期) — pure functions + unit tests

### Task 3: Auth & Authorization

**Files:**
- Create: `internal/transport/middleware/auth.go`
- Create: `internal/service/auth.go`
- Create: `internal/transport/handler/auth.go`

- [ ] 3.1 Implement JWT middleware: token generation, validation, refresh, middleware chain
- [ ] 3.2 Implement auth service: login, get current user info
- [ ] 3.3 Implement role-based middleware: admin/operator permission check
- [ ] 3.4 Implement user management API (admin): create user, delete user, list users
- [ ] 3.5 Create default admin account on first startup

### Task 4: Asset & Tenant Management

**Files:**
- Create: `internal/service/asset.go`
- Create: `internal/transport/handler/asset.go`
- Create: `internal/service/tenant.go`
- Create: `internal/transport/handler/tenant.go`

- [ ] 4.1 Implement Asset service + transport: CRUD API, search and pagination
- [ ] 4.2 Implement Tenant service + transport: CRUD API, search and pagination
- [ ] 4.3 Implement ExtraFields mechanism: settings define field list → UI dynamic rendering, data stored as JSON
- [ ] 4.4 Implement Settings API: GET/PUT `/api/settings/fields`, extended field config management

### Task 5: Contract Management

**Files:**
- Create: `internal/service/contract.go`
- Create: `internal/transport/handler/contract.go`

- [ ] 5.1 Implement Contract service: create (link asset+tenant, auto-calc total receivable, allow manual adjustment), query, update
- [ ] 5.2 Implement Contract transport: GET/POST `/api/contracts`, GET/PATCH `/api/contracts/:id`
- [ ] 5.3 Implement contract status auto-transition: based on TotalReceived vs TotalReceivable + current date
- [ ] 5.4 Implement template upload & field mapping API: GET/POST `/api/templates`, GET/PATCH `/api/templates/:id/mapping`

### Task 6: Payment & Arrears Collection

**Files:**
- Create: `internal/service/payment.go`
- Create: `internal/transport/handler/payment.go`
- Create: `internal/service/arrears.go`
- Create: `internal/transport/handler/arrears.go`

- [ ] 6.1 Implement Payment service: record payment, update contract TotalReceived, trigger calc engine recompute
- [ ] 6.2 Implement Payment transport: GET `/api/contracts/:id/payments`, POST `/api/contracts/:id/payments`
- [ ] 6.3 Implement Arrears service: daily auto-generate 5-level arrears list, query arrears history
- [ ] 6.4 Implement Arrears transport: GET `/api/arrears/lists`, GET `/api/arrears/history/:id`

### Task 7: Print Module

**Files:**
- Create: `internal/service/print.go`
- Create: `internal/transport/handler/print.go`
- Create: `internal/service/receiptbook.go`
- Create: `internal/transport/handler/receiptbook.go`

- [ ] 7.1 Implement contract PDF generation: load .docx template → field mapping replace placeholders → render PDF
- [ ] 7.2 Implement ID card copy PDF: 85.6mm×54mm scaled centered on A4, "仅用于租赁合同备案" watermark
- [ ] 7.3 Implement triple-copy receipt PDF: A4 portrait 3 equal sections (存根联/收据联/记账联), cut lines
- [ ] 7.4 Implement ReceiptBook management: CRUD API, print auto-increment sequence number
- [ ] 7.5 Implement Print transport: POST `/api/print/contract/:id`, POST `/api/print/receipt/:id`

### Task 8: Vue 3 Frontend Foundation

**Files:**
- Create: `frontend/` — Vue 3 project scaffold
- Create: `frontend/src/styles/variables.css`, global styles
- Create: `frontend/src/components/` — base components
- Create: `frontend/src/views/` — page views
- Create: `frontend/src/router/`, `frontend/src/stores/`, `frontend/src/api/`

- [ ] 8.1 Initialize Vue 3 project (Vite + Composition API + Pinia + Vue Router), configure build to embed output
- [ ] 8.2 Build Apple design system foundation: global style variables, base components (Button, Input, Card, StepIndicator)
- [ ] 8.3 Implement auth module: login page, token management, route guard, API interceptor
- [ ] 8.4 Implement layout frame: sidebar nav + main content area, 10 page route config

### Task 9: Three Core Entry Pages

**Files:**
- Create: `frontend/src/views/NewContract.vue`
- Create: `frontend/src/views/CollectRent.vue`
- Create: `frontend/src/views/ArrearsList.vue`

- [ ] 9.1 Implement "New Contract" page: step form (select asset → enter tenant → set contract → preview/print), asset/tenant dropdown search+create, OCR placeholder
- [ ] 9.2 Implement "Collect Rent" page: contract search + card list + payment modal, show "shortfall", one-click print receipt after payment
- [ ] 9.3 Implement "Arrears Collection" page: 5-level tab switch, urgency color coding, suggested actions per row

### Task 10: Admin Pages

**Files:**
- Create: `frontend/src/views/AssetList.vue`
- Create: `frontend/src/views/TenantList.vue`
- Create: `frontend/src/views/ContractList.vue`
- Create: `frontend/src/views/ReceiptBookList.vue`
- Create: `frontend/src/views/UserManagement.vue`
- Create: `frontend/src/views/Settings.vue`

- [ ] 10.1 Implement asset management: list+search+create/edit dialog, historical lease records
- [ ] 10.2 Implement tenant management: list+search+create/edit dialog, historical contract records
- [ ] 10.3 Implement contract management: list+status filter+detail (incl. payment records), template management
- [ ] 10.4 Implement receipt book management: list+create+status display
- [ ] 10.5 Implement user management (admin): user list+create+delete
- [ ] 10.6 Implement settings page: extended field configuration

### Task 11: Deployment & Integration

**Files:**
- Modify: `cmd/server/main.go` (embed)
- Create: `cmd/migrate/main.go`
- Create: `README.md`

- [ ] 11.1 Implement Go embed to bundle Vue build output, single binary
- [ ] 11.2 Implement data migration tool: SQLite → PostgreSQL
- [ ] 11.3 Write README and startup scripts
- [ ] 11.4 Integration tests: end-to-end verification of 3 core entry flows

### Task 12: Frontend Build Configuration

**Files:**
- Modify: `frontend/vite.config.ts`
- Modify: `cmd/server/main.go`

- [ ] 12.1 Configure Vite build: output to Go embed target directory, ensure no 404 on refresh
- [ ] 12.2 Go embed.FS embed dist/, static file serve + SPA fallback (non-API routes return index.html)
