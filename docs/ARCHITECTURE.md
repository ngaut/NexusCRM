# NexusCRM Architecture

## Vision

NexusCRM is a **Metadata-Driven CRM Platform**:
- **Go Backend**: Clean architecture, high performance
- **React Frontend**: Modern UI with Vite
- **TiDB Cloud**: Scalable MySQL-compatible database
- **Profile/Role System**: Salesforce-inspired permissions

Application structure (Schema), business logic (Flows), UI (Layouts), and security (Permissions) are stored as **metadata** in the database and interpreted at runtime.

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Frontend (React + Vite)                    │
│              http://localhost:5173                      │
│  Components → Services → API Client                     │
└──────────────────────┬──────────────────────────────────┘
                       │ REST API (JSON)
┌──────────────────────▼──────────────────────────────────┐
│          Backend (Go Clean Architecture)                │
│              http://localhost:3001                      │
│  ┌────────────────────────────────────────────┐        │
│  │ Interface Layer (REST Handlers)             │        │
│  │  auth, schema, data, formula, ui, approval  │        │
│  └──────────────────┬──────────────────────────┘        │
│  ┌──────────────────▼──────────────────────────┐        │
│  │ Application Layer (Business Logic)          │        │
│  │  MetadataService, QueryService, FlowEngine  │        │
│  └──────────────────┬──────────────────────────┘        │
│  ┌──────────────────▼──────────────────────────┐        │
│  │ Infrastructure Layer                        │        │
│  │  TiDB Connection (TLS), Event Bus           │        │
│  └──────────────────┬──────────────────────────┘        │
│                     │                           ┌───────▼──────────────┐
│                     │                           │ MCP Server           │
│                     │                           │ (Agent Tool Bus)     │
│                     │                           └──────────────────────┘
└──────────────────────┬──────────────────────────────────┘
                       │ MySQL Protocol + TLS
┌──────────────────────▼──────────────────────────────────┐
│                    TiDB Cloud                           │
│  - 40+ System Metadata Tables                          │
│  - User Business Tables                                │
└─────────────────────────────────────────────────────────┘
```

---

## Clean Architecture Layers

### Domain Layer (`internal/domain/`)
- **models/**: Business entities (User, Profile, Role)
- **events/**: Domain events for decoupling

### Application Layer (`internal/application/services/`)
- **MetadataService**: Schema management
- **QueryService**: Data retrieval with RLS
- **PermissionService**: RBAC and role hierarchy
- **FormulaEngine**: Excel-like expressions
- **FlowEngine**: Workflow automation

### Interface Layer (`internal/interfaces/rest/`)
- **auth_handler.go**: Login, logout, sessions
- **schema_handler.go**: Object/field metadata
- **data_handler.go**: CRUD operations
- **ui_handler.go**: Layouts, apps, tabs

### Infrastructure Layer (`internal/infrastructure/`)
- **database/tidb.go**: TiDB connection with TLS
- **events/**: Event bus implementation

---

## Permission System

### Profile (Required)
- Defines **what a user can do** (permissions)
- Controls CRUD on objects, field-level security
- Every user MUST have exactly one profile

### Role (Optional)
- Defines **whose data a user can see** (hierarchy)
- Tree structure: CEO > VP > Manager > Rep
- Higher roles inherit data access from lower roles

### Security Enforcement
1. **Object Permissions**: Create, Read, Edit, Delete, ViewAll, ModifyAll
2. **Field Permissions**: Readable, Editable
3. **Row-Level Security**: Ownership, role hierarchy, sharing rules

---

## System Tables (40+)

**Core Schema**: `_System_Object`, `_System_Field`, `_System_User`, `_System_Profile`, `_System_Role`

**Security**: `_System_ObjectPerms`, `_System_FieldPerms`, `_System_SharingRule`, `_System_RecordShare`

**UI**: `_System_App`, `_System_Layout`, `_System_Dashboard`, `_System_Tab`, `_System_ListView`

**Automation**: `_System_Flow`, `_System_Action`, `_System_Validation`, `_System_ApprovalProcess`

**Operations**: `_System_RecycleBin`, `_System_AuditLog`, `_System_Recent`

---

## Key Patterns

### Metadata-Driven
All UI, schema, and logic defined in database. No code changes needed for customization.

### Clean Architecture
Dependencies point inward: Interface → Application → Domain ← Infrastructure

### Event-Driven
EventBus decouples business logic. FlowEngine subscribes to domain events.

### React Portal for Modals
All modals use `createPortal(content, document.body)` with `z-[100]` for proper stacking.

---

## Agent Integration (MCP)

- **Protocol**: Model Context Protocol over HTTP
- **Endpoint**: `/mcp`
- **Authentication**: Standard JWT (agents are logical users)
- **Design**: 4-layer architecture (Foundation → Tool Bus → Runtime → Interaction)

See: [docs/architecture/agent_native_vision.md](./architecture/agent_native_vision.md)
