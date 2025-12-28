# NexusCRM Project


**Tech Stack:** React, TypeScript, Tailwind CSS, Go, TiDB Cloud, JWT
**Architecture:** 100% Metadata-Driven Platform as a Service (PaaS) with Modular Backend Services

## 🔒 Security First

NexusCRM includes enterprise-grade authentication and security:
- ✅ **JWT Authentication** with secure token management
- ✅ **bcrypt Password Hashing** (configurable rounds)
- ✅ **Strong Password Requirements** (8+ chars, uppercase, lowercase, number, special char)
- ✅ **Email Validation** with RFC 5322 compliance
- ✅ **SQL Injection Prevention** with strict parameterization and whitelist validation
- ✅ **System Admin-Only SQL Access** for enhanced security
- ✅ **Frontend-Backend Separation** with strict REST API boundaries
- ✅ **Modular Service Architecture** for maintainable security logic

See [docs/SECURITY.md](./docs/SECURITY.md) for details.

## 📂 Project Structure

```
/
├── frontend/                  # React Frontend
│   └── src/
│       ├── components/    # Reusable UI components
│       ├── constants/     # App constants
│       ├── contexts/      # React contexts
│       ├── core/          # Core framework logic
│       ├── infrastructure/# API & event services
│       ├── pages/         # Route components
│       ├── plugins/       # Action & UI plugins
│       ├── registries/    # Metadata registries
│       └── types.ts           # TypeScript Definitions
│
├── backend/                   # Go Backend (Clean Architecture)
│   ├── cmd/                   # Application Entry Points
│   │   └── server/            # HTTP Server
│   ├── internal/              # Private Application Code
│   │   ├── domain/            # Business Domain
│   │   │   ├── models/        # Domain Models
│   │   │   └── events/        # Domain Events
│   │   ├── application/       # Application Services
│   │   │   └── services/      # Business Logic (metadata, query, flow, auth)
│   │   ├── bootstrap/         # System Initialization
│   │   ├── interfaces/        # Interface Adapters
│   │   │   └── rest/          # REST API Handlers (auth, data, metadata, schema)
│   │   └── infrastructure/    # Infrastructure
│   │       ├── database/      # TiDB Connection
│   │       └── events/        # Event Bus
│   ├── pkg/                   # Public Packages
│   │   ├── auth/              # JWT Authentication & Password
│   │   ├── errors/            # Error Handling
│   │   ├── fieldtypes/        # Field Type Registry
│   │   ├── versioning/        # Optimized Version Management
│   │   └── constants/         # System Constants
│   └── scripts/               # Utility Scripts
│
├── tests/                     # Test Suites
│   ├── e2e/                   # Modular E2E Tests

│
├── shared/                    # Shared Definitions (Source of Truth)
│   └── constants/             # JSON constants for CodeGen
│
├── scripts/                   # Project-wide Utilities
│   └── generate-ts-constants.js # CodeGen script
│
├── mcp/                       # Model Context Protocol Server
│   ├── pkg/                   # MCP Logic
│   └── cmd/                   # Entry Points
│
└── docs/                      # Documentation
    ├── ARCHITECTURE.md        # System Architecture
    ├── SECURITY.md            # Security Features
```

## 🚀 Quick Start

1. **Install Dependencies**: `npm install`

2. **Configure Environment**:
   - **Root**: Copy `.env.example` to `.env`. Set TiDB credentials and `JWT_SECRET`.
   - **Optional**: 
     - `API_BASE_URL`: Override implementation URL (default: `http://localhost:3001` - used by internal MCP client)
     - `SKIP_ASSERTIONS=true`: Bypass strict startup checks (use with caution)
   - **Frontend**: Create `frontend/.env` and set `VITE_API_URL=http://localhost:3001`.

3. **Run Development**:
   ```bash
   # Run both Backend and Frontend concurrently (Recommended)
   npm run dev:full

   # OR Run individually:
   
   # Backend (Go)
   # Uses variables from root .env via package.json script
   npm run dev:server
   
   # Frontend (React + Vite)  
   npm run dev:client
   ```

4. **Default Credentials**:
   - Email: `admin@test.com`
   - Password: `Admin123!`

## 🏗 Architecture Highlights

### Clean Architecture Principles
- **Presentation → Application → Infrastructure** dependency flow
- **REST API boundaries** between frontend and backend
- **No direct database access** from the client
- **Proper separation of concerns** across all layers

### Metadata-Driven Platform
- **100% Metadata-Driven**: UI renders based on JSON configurations in database
- **42+ System Metadata Tables**: Complete platform configuration in database
- **Dynamic Schema Management**: Create objects and fields without code
- **Runtime Configuration**: Change behavior without deployment

### Security & Performance
- **Row-Level Security (RLS)**: Enforced in QueryEngine
- **Field-Level Security (FLS)**: Fine-grained permission control
- **JWT Authentication**: Secure token-based auth
- **Connection Pooling**: Optimized database connections

### Event-Driven Architecture
- **Event Bus**: Decouples UI actions from backend logic
- **Flow Engine**: Automated workflows
- **Formula Engine**: Excel-like calculated fields

### Modular Service Architecture
- **Decomposed Services**: Massive controllers broken down into focused, single-responsibility files (e.g., `schema_system_columns.go`, `permission_record_access.go`)
- **Strict Linting**: Comprehensive `errcheck` and static analysis enforcement
- **Maintainability**: Clear separation of business logic, validation, and persistence layers
- **Single Source of Truth**: `shared/constants/system.json` drives code generation for both Go and TypeScript, ensuring frontend-backend constant alignment.




## 📊 System Metadata Tables (40+ Total)

### Core Schema (9 tables)
- `_System_Object`, `_System_Field`, `_System_RecordType`, `_System_Relationship`, `_System_AutoNumber`
- `_System_Profile`, `_System_Role`, `_System_User`, `_System_Group`, `_System_GroupMember`

### Security & Permissions (6 tables)
- `_System_ObjectPerms`, `_System_FieldPerms`, `_System_SharingRule`, `_System_RecordShare`, `_System_Session`, `_System_PermissionSet`

### UI & Experience (14 tables)
- `_System_App`, `_System_Layout`, `_System_Dashboard`, `_System_Tab`, `_System_ListView`
- `_System_SetupPage`, `_System_UITheme`, `_System_UIComponent`, `_System_FieldRendering`
- `_System_NavigationMenu`, `_System_Limit`, `_System_Prompt`, `_System_Theme`

### Business Logic & Automation (11 tables)
- `_System_Flow`, `_System_Action`, `_System_ActionHandler`, `_System_Validation`
- `_System_FormulaFunction`, `_System_Transformation`, `_System_Webhook`, `_System_EmailTemplate`
- `_System_ApiEndpoint`, `_System_ApprovalProcess`, `_System_FieldDependency`

### Operations (5 tables)
- `_System_RecycleBin`, `_System_Log`, `_System_Recent`, `_System_AuditLog`, `_System_OutboxEvent`

## 📚 Documentation
- [ARCHITECTURE.md](./docs/ARCHITECTURE.md) - System architecture & design
- [USER_MANUAL.md](./docs/USER_MANUAL.md) - Usage instructions
- [SECURITY.md](./docs/SECURITY.md) - Security features & best practices
- [CONTRIBUTING.md](./CONTRIBUTING.md) - Development guide

## 🛠 Development Scripts

```bash
# Build
npm run build           # Build both client and server
npm run build:client    # Build frontend only
npm run build:server    # Build backend only

# Development
npm run dev             # Start both in dev mode
npm run dev:client      # Start frontend dev server
npm run dev:server      # Start backend dev server


# Testing
npm run test            # Run E2E tests
./backend/verify_custom_objects.sh # Verify Custom Object Lifecycle (Go)
```




