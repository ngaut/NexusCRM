# NexusCRM

**Tech Stack:** React, TypeScript, Tailwind CSS, Go, TiDB Cloud, JWT  
**Architecture:** 100% Metadata-Driven Platform with Clean Architecture

## Quick Start

```bash
# Install dependencies
npm install

# Configure environment
cp .env.example .env                     # Set TiDB credentials, JWT_SECRET
cp .env.example frontend/.env            # Set VITE_API_URL=http://localhost:3001

# Run development
npm run dev:full                         # Both backend + frontend
```

**Default Login:** `admin@test.com` / `Admin123!`

## Project Structure

```
/
├── frontend/                  # React + Vite
│   └── src/
│       ├── components/        # UI components
│       ├── core/              # Framework (constants, hooks)
│       ├── infrastructure/    # API services
│       ├── pages/             # Route components
│       ├── registries/        # Metadata registries
│       └── generated-schema.ts # Generated types
│
├── backend/                   # Go Clean Architecture
│   ├── cmd/server/            # Entry point
│   ├── internal/
│   │   ├── application/       # Business logic services
│   │   ├── interfaces/rest/   # REST handlers
│   │   └── infrastructure/    # TiDB, events
│   └── pkg/                   # Shared packages
│
├── shared/                    # Shared code generation
│   └── pkg/                   # Generated constants, models
│
├── mcp/                       # Model Context Protocol Server
├── tests/e2e/                 # E2E test suites
└── docs/                      # Documentation
```

## Key Features

- **Metadata-Driven**: Schema, UI, permissions stored as database config
- **40+ System Tables**: Complete platform configuration
- **Enterprise Security**: JWT auth, RLS, FLS, Profile/Role system
- **AI Assistant**: MCP-based agent integration

## Documentation

- [ARCHITECTURE.md](./docs/ARCHITECTURE.md) - System design
- [SECURITY.md](./docs/SECURITY.md) - Security features
- [USER_MANUAL.md](./docs/USER_MANUAL.md) - Usage guide
- [adding-fields.md](./docs/adding-fields.md) - Developer guide
- [PRD.md](./PRD.md) - Best practices & lessons learned

## Development

```bash
npm run dev:full        # Start both servers
npm run lint            # Run linting
npm run test            # E2E tests
make generate           # Regenerate code from system_tables.json
```

**Backend restart:**
```bash
./backend/restart-server.sh
```
