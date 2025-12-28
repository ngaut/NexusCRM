# NexusCRM Project


**Tech Stack:** React, TypeScript, Tailwind, Go, TiDB Cloud, JWT  
**Architecture:** 100% Metadata-Driven Platform as a Service (PaaS)

## 🔒 Security First

NexusCRM includes enterprise-grade authentication and security:
- ✅ **JWT Authentication** with secure token management
- ✅ **bcrypt Password Hashing** (configurable rounds)
- ✅ **Strong Password Requirements** (8+ chars, uppercase, lowercase, number, special char)
- ✅ **Email Validation** with RFC 5322 compliance
- ✅ **SQL Injection Prevention** with whitelist validation
- ✅ **System Admin-Only SQL Access** for enhanced security
- ✅ **Frontend-Backend Separation** with REST API boundaries

See [docs/SECURITY.md](./docs/SECURITY.md) for details.

## 📂 Project Structure

```
/
├── frontend/                  # React Frontend
│   └── src/
│       ├── components/        # UI Components
│       ├── pages/             # Page Views  
│       ├── hooks/             # Custom React Hooks
│       ├── contexts/          # React Contexts
│       ├── services/          # API Services
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
│   │   ├── interfaces/        # Interface Adapters
│   │   │   └── rest/          # REST API Handlers (auth, data, metadata, schema)
│   │   └── infrastructure/    # Infrastructure
│   │       ├── database/      # TiDB Connection
│   │       └── events/        # Event Bus
│   ├── pkg/                   # Public Packages
│   │   ├── auth/              # JWT Authentication & Password
│   │   ├── errors/            # Error Handling
│   │   ├── fieldtypes/        # Field Type Registry
│   │   └── constants/         # System Constants
│   └── scripts/               # Utility Scripts
│
├── tests/                     # Test Suites
│   ├── e2e/                   # Modular E2E Tests

│
└── docs/                      # Documentation
    ├── ARCHITECTURE.md        # System Architecture
    ├── SECURITY.md            # Security Features
    └── archive/               # Archived Docs
```

## 🚀 Quick Start

1. **Install Dependencies**: `npm install`

2. **Configure Environment**:
   - Copy `.env.example` to `.env` in both `frontend/` and `backend/`
   - Backend `.env`: Set TiDB connection (TIDB_HOST, TIDB_USER, TIDB_PASSWORD, TIDB_DATABASE)
   - Backend `.env`: Generate `JWT_SECRET`: `openssl rand -base64 32`
   - Frontend `.env`: Set `VITE_API_URL=http://localhost:3001`

3. **Run Development**:
   ```bash
   # Backend (Go)
   cd backend && go run cmd/server/main.go
   # Backend runs on http://localhost:3001
   
   # Frontend (React + Vite)  
   npm run dev
   # Frontend runs on http://localhost:5173
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
- **33 System Metadata Tables**: Complete platform configuration in database
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




## 📊 System Metadata Tables (33 Total)

### Core Metadata (8 tables)
- `_System_Object`, `_System_Field`, `_System_Profile`, `_System_ObjectPerms`, `_System_FieldPerms`, `_System_Role`, `_System_Session`, `_System_Config`

### UI Metadata (11 tables)
- `_System_Layout`, `_System_Dashboard`, `_System_App`, `_System_Tab`, `_System_SetupPage`, `_System_UITheme`, `_System_UIComponent`, `_System_FieldRendering`, `_System_NavigationMenu`, `_System_ListView`, `_System_Limit`

### Business Logic Metadata (9 tables)
- `_System_Flow`, `_System_Action`, `_System_ActionHandler`, `_System_Validation`, `_System_FormulaFunction`, `_System_Transformation`, `_System_Webhook`, `_System_EmailTemplate`, `_System_ApiEndpoint`

### Data Management (5 tables)
- `_System_SharingRule`, `_System_ProfileLayout`, `_System_RecycleBin`, `_System_Recent`, `_System_Log`

## 📚 Documentation

- [ARCHITECTURE.md](./docs/ARCHITECTURE.md) - System architecture & design
- [USER_MANUAL.md](./docs/USER_MANUAL.md) - Usage instructions
- [SECURITY.md](./docs/SECURITY.md) - Security features & best practices
- [DEBUGGING.md](./docs/DEBUGGING.md) - Debugging & distributed transaction tracing
- [REFACTORING_PLAN.md](./docs/REFACTORING_PLAN.md) - Metadata-driven PaaS roadmap

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

## Key Learnings
Key Learning for Future Tests: When creating apps with navigation items, always include:

json
{
    "id": "unique-id",
    "type": "object",  // or "page", "dashboard", "web"
    "label": "Display Name",
    "object_api_name": "api_name",  // for type: "object"
    "icon": "IconName"
}

## 📝 License

MIT
