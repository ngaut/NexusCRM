# E2E Test Suite

Modular, maintainable end-to-end tests for the NexusCRM backend API.

## Quick Start

```bash
# Run all tests
./runner.sh

# Run specific suites
./runner.sh auth metadata

# Run by suite number
./runner.sh 1 2 3

# List available suites
./runner.sh --list
```

## Structure

```
tests/e2e/
├── runner.sh          # Main test orchestrator
├── config.sh          # Environment configuration
├── lib/              # Shared utilities
│   ├── helpers.sh    # Test helpers & assertions
│   └── api.sh        # API request wrappers
└── suites/           # Test suites by domain
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

## Available Test Suites

1. **Infrastructure** - Health checks, process validation
2. **Authentication** - Login, protected endpoints, sessions
3. **Metadata** - Apps, schemas, object metadata
4. **CRUD** - Create, read, update, delete operations
5. **Search** - Global search, validation
6. **Formulas** - Formula engine, evaluation, functions
7. **Recycle Bin** - Soft delete, restore, purge
8. **Advanced Query** - Pagination, sorting, filtering
9. **Error Handling** - Edge cases, validation, error responses

## Configuration

Edit `config.sh` to customize:

```bash
export BASE_URL="http://localhost:3001"
export TEST_EMAIL="admin@test.com"
export TEST_PASSWORD="Admin123!"
```

## Running Individual Suites

Each suite can be run independently:

```bash
cd tests/e2e/suites
./02-auth.sh
```

## Writing New Tests

1. Create a new suite file in `suites/`
2. Follow the template:

```bash
#!/bin/bash
set +e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../config.sh"
source "$SCRIPT_DIR/../lib/helpers.sh"
source "$SCRIPT_DIR/../lib/api.sh"

SUITE_NAME="Your Suite Name"

run_suite() {
    section_header "$SUITE_NAME"
    test_your_feature
}

test_your_feature() {
    echo "Test: Your feature description"
    local response=$(api_get "/api/endpoint")
    assert_contains "$response" "expected" "Test passes"
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
```

3. Add to `AVAILABLE_SUITES` in `runner.sh`

## CI/CD Integration

```bash
# In your CI pipeline
cd tests/e2e
./runner.sh
exit_code=$?
# exit_code will be 0 if all tests pass, 1 if any fail
```

## Benefits

- ✅ **Modular** - Small, focused test files
- ✅ **Maintainable** - Clear separation of concerns
- ✅ **Flexible** - Run all or specific suites
- ✅ **Reusable** - Shared utilities, no duplication
- ✅ **CI/CD Ready** - Exit codes for automation
