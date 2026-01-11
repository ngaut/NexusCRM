# How to Add a Field to a System Table

## Overview

The `system_tables.json` file is the **single source of truth** for all system table schemas. When you add a field here, code generation automatically updates:

- Go constants (`shared/pkg/constants/z_generated_*.go`)
- Go models (`shared/pkg/models/z_generated.go`)
- TypeScript interfaces (`frontend/src/generated-schema.ts`)
- MCP types (`mcp/pkg/models/z_generated.go`)

## Step-by-Step Process

### Step 1: Edit `system_tables.json`

**Location**: `backend/internal/bootstrap/system_tables.json`

Add your new column to the `columns` array:

```json
{
    "tableName": "_System_User",
    "columns": [
        {"name": "id", "type": "VARCHAR(255)", "primaryKey": true},
        {"name": "email", "type": "VARCHAR(255)", "nullable": false},
        {"name": "phone", "type": "VARCHAR(40)", "nullable": true}
    ]
}
```

### Column Properties

| Property | Required | Description |
|----------|----------|-------------|
| `name` | ✅ | Column name (snake_case) |
| `type` | ✅ | SQL type (e.g., `VARCHAR(255)`, `TINYINT(1)`) |
| `nullable` | ❌ | Allows NULL values (default: false) |
| `primaryKey` | ❌ | Primary key column |
| `unique` | ❌ | Enforce uniqueness |
| `default` | ❌ | Default value |
| `logicalType` | ❌ | Type hint: `Lookup`, `Password` |
| `referenceTo` | ❌ | Target table for lookups |

### Step 2: Run Code Generation

```bash
make generate
```

### Step 3: Verify Build

```bash
make build
```

### Step 4: Database Migration

For production, run SQL migration:
```sql
ALTER TABLE `_System_User` ADD COLUMN `phone` VARCHAR(40);
```

For development, wipe and recreate:
```bash
cd backend && go run scripts/wipe_db.go
./restart-server.sh
```

## Generated Output Example

Adding `phone` to `_System_User` generates:

**Go Constant** (`shared/pkg/constants/z_generated_fields.go`):
```go
FieldSysUser_Phone = "phone"
```

**TypeScript** (`frontend/src/generated-schema.ts`):
```typescript
phone?: string;
```

## Common Types

| Purpose | SQL Type | Go Type | TypeScript |
|---------|----------|---------|------------|
| ID | `VARCHAR(255)` | `string` | `string` |
| Text | `VARCHAR(100)` | `string` | `string` |
| Boolean | `TINYINT(1)` | `bool` | `boolean` |
| Number | `INT` | `int` | `number` |
| DateTime | `DATETIME` | `time.Time` | `string` |

## Using Constants

```go
// ✅ Compile-time safe
query := fmt.Sprintf("SELECT %s FROM %s",
    constants.FieldSysUser_Email,
    constants.TableUser,
)
```

## CI Verification

`make verify-generated` fails if generated files are out of sync.
