# How to Add a Field to a System Table

This guide explains how to add a new field to a system table using the Single Source of Truth (SSOT) architecture.

## Overview

The `system_tables.json` file is the **single source of truth** for all system table schemas. When you add a field here, code generation automatically updates:

- Go table/field constants (`backend/pkg/constants/z_generated_*.go`)
- Go struct types (`backend/internal/domain/models/z_generated.go`)
- TypeScript interfaces (`frontend/src/generated-schema.ts`)
- MCP types (`mcp/pkg/models/z_generated.go`)

## Step-by-Step Process

### Step 1: Edit `system_tables.json`

Location: `backend/internal/bootstrap/system_tables.json`

Find the table you want to modify and add your new column to the `columns` array:

```json
{
    "tableName": "_System_User",
    "columns": [
        // ... existing columns ...
        {
            "name": "phone",
            "type": "VARCHAR(40)",
            "nullable": true
        }
    ]
}
```

### Column Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `name` | string | ✅ | Column name (snake_case) |
| `type` | string | ✅ | SQL type (e.g., `VARCHAR(255)`, `TINYINT(1)`, `DATETIME`) |
| `nullable` | boolean | ❌ | If true, allows NULL values (default: false) |
| `primaryKey` | boolean | ❌ | If true, this is the primary key |
| `unique` | boolean | ❌ | If true, enforces uniqueness |
| `default` | string | ❌ | Default value (e.g., `"0"`, `"CURRENT_TIMESTAMP"`) |
| `logicalType` | string | ❌ | High-level type hint (e.g., `Lookup`, `Password`) |
| `referenceTo` | string | ❌ | For lookups, the target table name |
| `isNameField` | boolean | ❌ | If true, used as display name for records |

### Step 2: Run Code Generation

```bash
make generate
```

This regenerates all constant files, struct definitions, and TypeScript types.

### Step 3: Verify Everything Compiles

```bash
make build
```

This runs:
- `go build ./...` (backend)
- `go build ./...` (mcp)
- `npm run lint` (frontend TypeScript check)

### Step 4: Add Database Migration (If Needed)

If you're adding a column to an existing production database, you'll need to run a migration:

```sql
ALTER TABLE `_System_User` ADD COLUMN `phone` VARCHAR(40);
```

For development, you can wipe the database and let it recreate:

```bash
./scripts/wipe_db.sh
./scripts/restart-server.sh
```

## Example: Adding `phone` to `_System_User`

### Before (JSON)
```json
{
    "tableName": "_System_User",
    "columns": [
        {"name": "id", "type": "VARCHAR(255)", "primaryKey": true},
        {"name": "username", "type": "VARCHAR(255)", "nullable": false},
        {"name": "email", "type": "VARCHAR(255)", "nullable": false}
    ]
}
```

### After (JSON)
```json
{
    "tableName": "_System_User",
    "columns": [
        {"name": "id", "type": "VARCHAR(255)", "primaryKey": true},
        {"name": "username", "type": "VARCHAR(255)", "nullable": false},
        {"name": "email", "type": "VARCHAR(255)", "nullable": false},
        {"name": "phone", "type": "VARCHAR(40)", "nullable": true}
    ]
}
```

### Generated Output

**Go Constant** (`z_generated_fields.go`):
```go
const FieldSysUser_Phone = "phone"
```

**Go Struct** (`z_generated.go`):
```go
type GenSystemUser struct {
    // ...
    Phone *string `json:"phone,omitempty"`
}
```

**TypeScript** (`generated-schema.ts`):
```typescript
export interface SystemUser {
    // ...
    phone?: string;
}
```

## Common Field Types

| Purpose | SQL Type | Go Type | TypeScript Type |
|---------|----------|---------|-----------------|
| ID/UUID | `VARCHAR(255)` | `string` | `string` |
| Short text | `VARCHAR(100)` | `string` | `string` |
| Long text | `TEXT` | `string` | `string` |
| Boolean | `TINYINT(1)` | `bool` | `boolean` |
| Integer | `INT` | `int` | `number` |
| Decimal | `DECIMAL(18,4)` | `float64` | `number` |
| Date/time | `DATETIME` | `time.Time` | `string` |
| JSON data | `JSON` | `json.RawMessage` | `Record<string, unknown>` |
| Lookup FK | `VARCHAR(255)` + `logicalType: Lookup` | `*string` | `string?` |

## Using Generated Constants

Prefer using the generated table-specific constants for compile-time safety:

```go
// ✅ Good - Compile-time safe
query := fmt.Sprintf("SELECT %s, %s FROM %s",
    constants.FieldSysUser_Email,
    constants.FieldSysUser_Phone,
    constants.TableUser,
)

// ❌ Avoid - Runtime error risk
query := fmt.Sprintf("SELECT email, phonee FROM _System_User") // typo!
```

## CI Verification

The CI pipeline runs `make verify-generated` which will **fail** if:
1. You modified `system_tables.json`
2. But didn't run `make generate`

This ensures generated files are always in sync.

## Troubleshooting

### "Column not found" errors after adding field
- Run database migration or wipe database
- Ensure you ran `make generate`

### TypeScript type errors
- Run `npm run lint` to check for issues
- Ensure generated-schema.ts is imported correctly

### Constant not found in Go
- Ensure you ran `make generate`
- Check the constant name format: `FieldSys{Table}_{FieldName}`
