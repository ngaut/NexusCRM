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

## Supported Field Types (from `fieldTypes.json`)

| Logical Type | SQL Type | Description |
|--------------|----------|-------------|
| `Text` | `VARCHAR(255)` | Short text (Name, Title) |
| `TextArea` | `TEXT` | Multi-line text (Address) |
| `LongTextArea` | `LONGTEXT` | Large text content |
| `RichText` | `LONGTEXT` | HTML content |
| `Number` | `DECIMAL(18,2)` | Numeric with decimals |
| `Currency` | `DECIMAL(18,2)` | Monetary values |
| `Percent` | `DECIMAL(5,2)` | Percentage (0.00-100.00) |
| `Date` | `DATE` | Date only |
| `DateTime` | `DATETIME` | Date and Time |
| `Boolean` | `BOOLEAN` | Checkbox (True/False) |
| `Picklist` | `VARCHAR(255)` | Single select option |
| `Email` | `VARCHAR(255)` | Email with validation |
| `Phone` | `VARCHAR(40)` | Phone number |
| `Url` | `VARCHAR(1024)` | Web link |
| `Lookup` | `VARCHAR(36)` | Reference IDs (Foreign Key) |
| `JSON` | `JSON` | Structured data |

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
