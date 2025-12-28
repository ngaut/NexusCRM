# NexusCRM Master Roadmap: The "Salesforce-like" Vision

This document outlines the strategic roadmap to evolve NexusCRM into a comprehensive, metadata-driven Platform-as-a-Service (PaaS) comparable to Salesforce.

**Last Updated: Dec 29, 2025**

## 1. Core Platform Architecture (The "Kernel")
The heart of the system. Flexible, metadata-driven data model.

### 1.1 Object Model (95% Complete)
- [x] **Custom Objects**: Create tables on the fly (`_System_Object`, `_System_Field`).
- [x] **Field Types**: Text, Number, Currency, Date, DateTime, Boolean, Picklist, Lookup.
- [x] **Advanced Relationships**:
    - [x] **Master-Detail**: Hard link with cascading delete and rollup summaries.
    - [x] **Junction Objects**: Many-to-Many relationships (E2E Suite 11 tested).
    - [x] **Polymorphic Lookups**: Look up to multiple object types (E2E Suite 12 tested).
- [x] **Roll-up Summary Fields**: Aggregate data (Sum, Max, Min, Count) from child records to parent.
- [x] **Formula Fields**: Calculated fields (Application-level, read-time evaluation).
- [ ] **Big Objects**: High-volume, immutable data storage for logs/audit.
- [x] **Modular Backend**: Decomposed monolithic services into 100% maintainable, discrete components (Completed Dec 2025).

### 1.2 Identity & Security (The "Shield") (80% Complete)
Granular control over who sees what.

- [x] **Authentication**: JWT, Session management.
- [x] **Organization-Wide Defaults (OWD)**: Private/Public Read/Public Write.
- [x] **Profiles**: Object (CRUD) and Field (FLS) permissions.
- [x] **Permission Sets**: Additive permissions beyond Profile (E2E Suite 13 tested).
- [x] **Role Hierarchy**:
    - [x] **Schema**: `_System_Role` table and `User.role_id` exist.
    - [x] **Logic**: Implement "Manager" visibility in `PermissionService`.
- [x] **Sharing Rules**:
    - [x] **Schema**: `_System_SharingRule` table.
    - [x] **Logic**: Engine to evaluate criteria and grant access based on role membership.
- [x] **Manual Sharing**: Share specific records with users/groups (`_System_RecordShare`).
- [x] **Teams**: Account/Opportunity Teams (`_System_TeamMember`).

## 2. Automation & Logic (The "Brain") (60% Complete)
Business logic execution without code.

- [x] **Triggers (Flows)**:
    - [x] **Engine**: `FlowExecutor` supports Before/After CRUD, Field Updates (`updateRecord`), Creation (`createRecord`), Email (`sendEmail`), Webhooks (`callWebhook`).
    - [x] **Expressions**: Formula-based conditions and field value assignment (`=TODAY()`).
    - [x] **CRUD API**: List, Create, Get, Update, Delete flows (E2E Suite 20 tested).
- [ ] **Advanced Flows**:
    - [ ] **Screen Flows**: Multi-step wizard UI for users.
    - [x] **Auto-Launched Flows**: REST-invocable logic chains (`POST /api/flows/:id/execute`).
- [x] **Approval Processes**: State machine for record approval (Submit → Approve/Reject).
- [x] **Validation Rules**: Formula-based error prevention.
- [x] **Custom Actions**: Execute server-side logic (E2E Suite 14 tested).

## 3. User Interface (The "Lightning Experience") (70% Complete)
Dynamic, component-based UI.

- [x] **App Builder**: Define Apps (collections of Tabs).
- [x] **Page Layout Editor**:
    - [x] Drag-and-drop fields, sections, related lists.
    - [x] **Backend Performance Optimization**: In-Memory Metadata Caching (~445ms latency).
- [x] **Layout Editor Reliability**: Fixed Drag-and-Drop flakiness.
- [/] **Dynamic Forms**:
    - [x] Section Visibility (Rule Engine).
    - [ ] Field Visibility (Deferred).
- [x] **List Views**: Filterable grids, Kanban board.
- [x] **Split View**: List on left, detail on right (`SplitViewContainer.tsx`).
- [x] **Mass Inline Edit**: Bulk edit multiple records (`BulkEditModal.tsx`).
- [x] **Global Search**: Federated search across all objects with relevancy ranking.
- [ ] **Console View**: Multi-tab workspace for support agents.

## 4. Analytics & Intelligence (The "Vision") (95% Complete)
Turning data into insights.

- [x] **Query Engine**:
    - [x] **Raw SQL**: `AnalyticsHandler` allows admins to run raw SQL.
    - [x] **SQL Chart Widget**: Execute queries and visualize as charts.
    - [x] **CSV Export**: Export query results from SQL Chart widget.
- [x] **Dashboard Builder**:
    - [x] **Schema**: `_System_Dashboard` exists.
    - [x] **Editor**: `StudioDashboardEditor` with drag-and-drop `react-grid-layout`.
    - [x] **Widget Registry**: 9 widget types (Metric, Charts, Kanban, RecordList, SQL).
    - [x] **Property Inspector**: Edit widget configs via side panel.
    - [x] **Dashboard Viewer**: Refresh button with "Updated" timestamp.
- [x] **List View Charts**: Quick analytics panel above record lists (collapsible).

## 5. Development & Integration (The "Gateway") (70% Complete)
Extensibility for developers.

- [x] **Auto-REST API**: JSON API for every object automatically.
- [ ] **Bulk API**: Async processing for millions of records (CSV import/export).
- [ ] **Streaming API**: Server-Sent Events (SSE) or WebSockets for record changes.
- [ ] **Metadata API**: Export/Import configuration (XML/JSON) for deployments.
- [ ] **Sandboxes**: Mechanism to clone Production → Staging environment.
- [ ] **Marketplace**: Installable "Packages" (bundles of metadata).

## 6. Standard Applications (The "Sales & Service Cloud") (40% Complete)
Out-of-the-box functionality.

- [x] **Foundation**: Accounts, Contacts, Leads, Opportunities, Cases, Tasks.
- [/] **Sales Cloud**:
    - [x] **Leads**: Lead object with status workflow (E2E Suite 16 tested).
    - [x] **Opportunities**: Sales stages, forecasting fields.
    - [x] **Lead Conversion**: Creates Account + Contact + Opportunity (`LeadConvertPlugin`).
    - [ ] **Products & Pricebooks**: Line items, Multi-currency.
- [/] **Service Cloud**:
    - [x] **Cases**: Case object with status workflow (E2E Suite 17 tested).
    - [ ] **Email-to-Case**: Inbound email parsing.
    - [ ] **Knowledge Base**: Articles, Categories.
- [ ] **Activity Management**: Calendar, Recurring Events.

## 7. Testing Infrastructure (100% Complete)
Comprehensive E2E test coverage.

- [x] **E2E Test Suites**: 42 suites covering all major features
- [x] **Helper Libraries**: `schema_helpers.sh`, `test_data.sh` for DRY tests
- [x] **Domain Scenarios**: HR, E-Commerce, Real Estate, Healthcare, Education, Jira
- [x] **Performance Tests**: Bulk operations, concurrent workloads (Suite 27-28)
- [x] **Validation Tests**: Edge cases, API response formats (Suite 29-30)

---

## Execution Strategy

### Phase 1: Foundation (Completed)
- Solidified Schema, Metadata, and Basic Security.
- **Achieved**: Reliable CRUD and App Platform with 42 E2E test suites.

### Phase 2: The "Killer Features" (Completed)
- [x] Phase 2.1: Schema Manager Decomposition
- [x] Phase 2.2: Backend Service Decomposition
- [x] Documentation Audit & Cleanup (Dec 2025)
- [x] Deep Scrub: Removed non-existent Agent/Queue tables from docs
- [x] Project Structure: Added `shared/` and `scripts/` to README
- [x] Auxiliary Docs: Synced `.env.example`, `CONTRIBUTING.md`, `package.json`
- **Advanced Relationships**: Master-Detail & Rollups ✅
- **Sharing Engine**: Role hierarchy and sharing rules ✅
- **Agent-Native**: MCP Server Implemented ✅
- **Done**: Approval Processes, Lead Conversion.

### Phase 3: The "Experience"
- **Visual Flow Builder**: Empower non-coders.
- **Page Layout Editor**: Flexible UI ✅
- **Global Search**: Implemented ✅
- **Next**: Screen Flows.

### Phase 4: The Ecosystem
- **Packages & Marketplace**: Export apps.
- **Mobile App**: Native/PWA support.
- **Bulk/Streaming APIs**: Enterprise data integration.
