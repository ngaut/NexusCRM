# NexusCRM Architecture

## 1. Vision & Philosophy

NexusCRM is a **Metadata-Driven CRM Platform** built with:
- **Go Backend**: High-performance, clean architecture
- **React Frontend**: Modern UI with Vite
- **TiDB Cloud**: Scalable MySQL-compatible database
- **Profile/Role System**: Salesforce-inspired permissions

The core philosophy is that application structure (Schema), business logic (Flows), user interface (Layouts), and security (Permissions) are stored as **metadata** in the database. The system interprets this metadata at runtime to construct UI and execute logic dynamically.

**For user instructions, see [USER_MANUAL.md](./USER_MANUAL.md).**

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Frontend (React + Vite)                │
│                   http://localhost:5173                  │
│                                                          │
│  Components → Services → API Client                    │
│   - UI Components                                       │
│   - Custom Hooks                                        │
│   - Context Providers                                   │
└──────────────────────┬──────────────────────────────────┘
                       │ REST API (JSON)
┌──────────────────────▼──────────────────────────────────┐
│              Backend (Go Clean Architecture)            │
│                   http://localhost:3001                  │
│                                                          │
│  ┌────────────────────────────────────────────┐        │
│  │ Interface Layer (REST Handlers)             │        │
│  │  - auth_handler.go                          │        │
│  │  - schema_handler.go                        │        │
│  │  - data_handler.go                          │        │
│  └──────────────────┬──────────────────────────┘        │
│  ┌──────────────────▼──────────────────────────┐        │
│  │ Application Layer (Business Logic)          │        │
│  │  - MetadataService                          │        │
│  │  - QueryService                             │        │
│  │  - PermissionService                        │        │
│  │  - FormulaEngine                            │        │
│  └──────────────────┬──────────────────────────┘        │
│  ┌──────────────────▼──────────────────────────┐        │
│  │ Domain Layer (Models & Events)              │        │
│  │  - Business Models                          │        │
│  │  - Domain Events                            │        │
│  └──────────────────┬──────────────────────────┘        │
│  ┌──────────────────▼──────────────────────────┐        │
│  │ Infrastructure Layer                        │        │
│  │  - TiDB Connection (with TLS)               │        │
│  │  - Event Bus                                │        │
│  └────────────────────────────────────────────┘        │
└──────────────────────┬──────────────────────────────────┘
                       │ MySQL Protocol + TLS
┌──────────────────────▼──────────────────────────────────┐
│                    TiDB Cloud                           │
│              (MySQL-Compatible Database)                 │
│                                                          │
│  - 33 System Metadata Tables                           │
│  - User Business Tables                                 │
│  - ACID Transactions                                    │
└─────────────────────────────────────────────────────────┘
```

---

## 3. Clean Architecture Layers

### 3.1 Domain Layer (`internal/domain/`)
**Purpose**: Core business logic and models

- **models/**: Business entities (User, Profile, Role, etc.)
- **events/**: Domain events for decoupling

**Key Principle**: No dependencies on outer layers

### 3.2 Application Layer (`internal/application/services/`)
**Purpose**: Use cases and business operations

**Core Services**:
- **MetadataService**: Schema management, delegates to `SchemaDefaults` for system field injection.
- **QueryService**: Data retrieval with RLS enforcement
- **PermissionService**: RBAC and role hierarchy
- **FormulaEngine**: Excel-like expression evaluator
- **FlowEngine**: Workflow automation
- **EventBus**: Pub/sub for decoupled logic
- **SchemaManager**: Handles physical DDL operations
- **UIMetadataService**: Manages layouts, dashboards, and apps

### 3.3 Interface Layer (`internal/interfaces/rest/`)
**Purpose**: HTTP handlers and API contracts

**Main Handlers**:
- **auth_handler.go**: Login, logout, session management
- **schema_handler.go**: Schema discovery, object/field metadata
- **data_handler.go**: CRUD operations on records
- **formula_handler.go**: Formula evaluation
- **ui_handler.go**: UI metadata (layouts, apps, tabs)

**Pattern**: Each handler delegates to application services

### 3.4 Infrastructure Layer (`internal/infrastructure/`)
**Purpose**: External dependencies and technical concerns

- **database/tidb.go**: TiDB Cloud connection with TLS
- **events/**: Event bus implementation

---

## 4. Permission System (Salesforce Pattern)

### 4.1 Profile vs Role

**Profile** (Required):
- Defines **what a user can do** (permissions)
- Examples: `system_admin`, `standard_user`
- Controls CRUD operations on objects
- Controls field-level security
- Every user MUST have exactly one profile

**Role** (Optional):
- Defines **whose data a user can see** (hierarchy)
- Role hierarchy enables managers to see subordinate data
- Not all users need roles (e.g., individual contributors)
- Example: CEO > VP Sales > Regional Manager > Sales Rep

**Database Schema**:
```sql
_System_User:
  ProfileId varchar(255) NOT NULL  -- Required: Permissions
  RoleId    varchar(255) NULL       -- Optional: Hierarchy
```

**API Response**:
```json
{
  "id": "user-id",
  "email": "user@example.com",
  "profileId": "system_admin",  // Always present
  "roleId": null                 // Always present, can be null
}
```

### 4.2 Permission Evaluation

1. **Profile Permissions** (`_System_ObjectPerms`, `_System_FieldPerms`)
   - Object-level: Create, Read, Edit, Delete, ViewAll, ModifyAll
   - Field-level: Readable, Editable

2. **Role Hierarchy** (`_System_Role`)
   - Tree structure
   - Higher roles see data owned by lower roles
   - Computed in PermissionService

3. **Sharing Rules** (`_System_SharingRule`)
   - Criteria-based access extensions
   - Example: "If Industry='Tech', share with Sales Team"

---

## 5. Data Model

### 5.1 System Metadata Tables (Verified)

**Core Metadata**:
- `_System_Object`, `_System_Field`, `_System_RecordType`
- `_System_Profile`, `_System_Role`, `_System_User`, `_System_Session`
- `_System_ObjectPerms`, `_System_FieldPerms`
- `_System_Config`, `_System_AutoNumber`

**UI Metadata**:
- `_System_App`, `_System_Layout`, `_System_Dashboard`
- `_System_ListView`, `_System_SetupPage`
- `_System_UIComponent`

**Business Logic**:
- `_System_Flow`, `_System_Action`
- `_System_Validation`, `_System_FieldDependency`

**Data Management**:
- `_System_SharingRule` (Schema only)
- `_System_RecycleBin`, `_System_Recent`, `_System_Log`

### 5.2 Business Tables
- **Standard Objects**: Account, Contact, Opportunity, Lead, Task
- **Custom Objects**: User-defined (suffix `__c`)

---

## 6. Security Architecture

### 6.1 Authentication
- **JWT Tokens**: Secure, stateless authentication
- **bcrypt Password Hashing**: Industry-standard (12 rounds)
- **Session Management**: Stored in `_System_Session`

### 6.2 Row-Level Security (RLS)
**Enforced in QueryService**:
1. **Object-Level Permissions**: Read/Create/Edit/Delete permissions checked per object.
2. **Field-Level Visibility**: Fields hidden based on profile permissions.
3. **Soft Deletion**: Automatically excludes deleted records (`is_deleted=0`).
*Note: Advanced Sharing Rules and Role Hierarchy visibility are planned for future release.*

### 6.3 Field-Level Security (FLS)
**Enforced at API layer**:
- Read permissions: Controls visibility
- Edit permissions: Controls mutability
- Checked in both query and persistence operations

### 6.4 Metadata Persistence & Hydration
**Reliability Pattern**:
To ensure data integrity, the system uses a **hydration pattern** for all object creation.
- **Schema Defaults**: When a new custom object is defined (e.g., `Project__c`), `schema_defaults.go` automatically injects mandatory system fields:
    - `id` (UUID, Primary Key)
    - `created_date`, `created_by_id`
    - `last_modified_date`, `last_modified_by_id`
    - `owner_id`, `is_deleted`
- **Atomic Operations**: Metadata is registered in `_System_Object` and `_System_Field` transactional tables simultaneously with physical table creation (`CREATE TABLE`). This prevents "orphan" tables or "ghost" metadata.

---

## 7. Testing Architecture

### 7.1 Modular E2E Tests (`tests/e2e/`)
```
tests/e2e/
├── runner.sh              # Test orchestrator
├── config.sh              # Environment config
├── lib/
│   ├── helpers.sh         # Test utilities
│   └── api.sh             # API wrappers
└── suites/
    ├── 01-infrastructure.sh
    ├── 02-auth.sh
    ├── 03-metadata.sh
    ├── 04-crud.sh
    ├── 05-search.sh
    ├── 06-formulas.sh
    ├── 07-recyclebin.sh
    ├── 08-advanced-query.sh
    └── 09-error-handling.sh
```

**Run Tests**:
```bash
cd tests/e2e && ./runner.sh          # All tests
cd tests/e2e && ./runner.sh auth    # Specific suite
```

### 7.2 Role Implementation Tests
`tests/comprehensive_role_tests.sh` - Validates Profile/Role system end-to-end

---

## 8. Development

### 8.1 Project Structure
```
nexuscrm/
├── frontend/              # React + TypeScript + Vite
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── contexts/
│   │   ├── services/
│   │   └── types.ts
│   └── vite.config.ts
│
├── backend/               # Go + Clean Architecture
│   ├── cmd/server/        # Entry point
│   ├── internal/
│   │   ├── domain/
│   │   ├── application/
│   │   ├── interfaces/
│   │   └── infrastructure/
│   ├── pkg/
│   │   ├── auth/          # JWT utilities
│   │   ├── errors/
│   │   └── constants/
│   └── go.mod
│
└── tests/                 # Test suites
    ├── e2e/
    └── comprehensive_role_tests.sh
```

### 8.2 Running Locally
```bash
# Backend (Terminal 1)
cd backend
go run cmd/server/main.go
# → http://localhost:3001

# Frontend (Terminal 2)
npm run dev
# → http://localhost:5173
```

### 8.3 Environment Configuration
**Backend** (`.env`):
```bash
TIDB_HOST=gateway01.us-west-2.prod.aws.tidbcloud.com
TIDB_PORT=4000
TIDB_USER=<your-user>
TIDB_PASSWORD=<your-password>
TIDB_DATABASE=nexuscrm
JWT_SECRET=<generate-with-openssl-rand>
```

**Frontend** (`.env`):
```bash
VITE_API_URL=http://localhost:3001
```

---

## 9. Key Design Patterns

### 9.1 Metadata-Driven
- All UI, schema, and logic defined in database
- No code changes needed for customization
- Runtime interpretation of metadata

### 9.2 Clean Architecture
- Dependencies point inward (Domain ← Application ← Interface ← Infrastructure)
- Business logic independent of frameworks
- Testable, maintainable, scalable

### 9.3 Event-Driven
- EventBus decouples business logic
- FlowEngine subscribes to events
- Atomic operations with transactions

### 9.4 Security-First
- RLS enforced in QueryService
- FLS checked at API boundaries
- Profile/Role separation (Salesforce pattern)

### 9.5 Agent-Native Architecture (Planned)
- **4-Layer Design**: Foundation (L1) → Dynamic Tool Bus (L2) → Runtime (L3) → Interaction (L4).
- **Protocol**: Model Context Protocol (MCP) for tool exposure.
- **Philosophy**: Agents are logical users subject to standard permissions but empowered with dynamic discovery.

See detailed design documents:
- [Agent-Native Vision (v3)](architecture/agent_native_vision.md)
- [Platform Technical Design](architecture/agent_platform_design.md)
- [UI/UX Interaction & Design](architecture/agent_ui_design.md)


---


