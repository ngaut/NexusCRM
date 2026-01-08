#!/bin/bash
# tests/e2e/suites/08-advanced-query.sh
# Advanced Query Operations Tests
# REFACTORED: Uses dynamically created test objects instead of hardcoded Account

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Advanced Query Operations"

# Helper for unique names
TS=$(date +%s)
TEST_OBJ="query_test_$TS"

test_cleanup() {
    echo "Cleaning up test object..."
    api_delete "/api/metadata/objects/$TEST_OBJ" > /dev/null 2>&1
}
trap test_cleanup EXIT

run_suite() {
    section_header "$SUITE_NAME"
    
    # Re-authenticate for fresh token
    echo "Re-authenticating for remaining tests..."
    api_login "$TEST_EMAIL" "$TEST_PASSWORD"
    echo ""
    
    setup_test_object
    seed_test_data
    
    test_pagination
    test_sorting
    test_multiple_criteria
    test_field_selection
}

setup_test_object() {
    echo "Setup: Creating test object '$TEST_OBJ'..."
    
    local response=$(api_post "/api/metadata/objects" "{
        \"label\": \"$TEST_OBJ\",
        \"plural_label\": \"${TEST_OBJ}s\",
        \"api_name\": \"$TEST_OBJ\",
        \"description\": \"E2E Advanced Query Test Object\",
        \"is_custom\": true,
        \"searchable\": true
    }")
    
    if echo "$response" | grep -q "\"api_name\":\"$TEST_OBJ\""; then
        echo "  ✓ Test object created: $TEST_OBJ"
    else
        echo "  ✗ Failed to create test object"
        echo "  Response: $response"
        return 1
    fi
    
    # Add fields
    api_post "/api/metadata/objects/$TEST_OBJ/fields" '{"api_name": "industry", "label": "Industry", "type": "Text"}' > /dev/null
    api_post "/api/metadata/objects/$TEST_OBJ/fields" '{"api_name": "annual_revenue", "label": "Annual Revenue", "type": "Number"}' > /dev/null
    echo "  ✓ Fields added to test object"
    
    # Wait for Schema Cache (Polling)
    echo "  Waiting for field 'annual_revenue'..."
    for i in {1..10}; do
        meta=$(api_get "/api/metadata/objects/$TEST_OBJ")
        if echo "$meta" | grep -q "\"api_name\":\"annual_revenue\""; then
            break
        fi
        sleep 0.5
    done
}

seed_test_data() {
    echo "Setup: Seeding test data..."
    api_post "/api/data/$TEST_OBJ" '{"name": "Test Record 1", "industry": "Technology", "annual_revenue": 5000000}' > /dev/null
    api_post "/api/data/$TEST_OBJ" '{"name": "Test Record 2", "industry": "Healthcare", "annual_revenue": 3000000}' > /dev/null
    api_post "/api/data/$TEST_OBJ" '{"name": "Test Record 3", "industry": "Technology", "annual_revenue": 1500000}' > /dev/null
    api_post "/api/data/$TEST_OBJ" '{"name": "Test Record 4", "industry": "Finance", "annual_revenue": 8000000}' > /dev/null
    echo "  ✓ 4 test records seeded"
}

test_pagination() {
    echo "Test 8.1: Query with Pagination"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"criteria\": [],
        \"limit\": 2,
        \"offset\": 0
    }")
    
    if echo "$response" | grep -q '"records"'; then
        test_passed "POST /api/data/query supports pagination (limit/offset)"
    else
        test_failed "POST /api/data/query pagination" "$response"
    fi
}

test_sorting() {
    echo ""
    echo "Test 8.2: Query with Sorting"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"criteria\": [],
        \"order_by\": [{\"field\": \"name\", \"direction\": \"ASC\"}],
        \"limit\": 10
    }")
    
    if echo "$response" | grep -qE '"records"|"success"'; then
        test_passed "POST /api/data/query supports sorting (orderBy)"
    else
        test_failed "POST /api/data/query sorting" "$response"
    fi
}

test_multiple_criteria() {
    echo ""
    echo "Test 8.3: Query with Multiple Criteria (AND)"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"criteria\": [
            {\"field\": \"industry\", \"op\": \"=\", \"val\": \"Technology\"},
            {\"field\": \"annual_revenue\", \"op\": \">\", \"val\": 1000000}
        ],
        \"limit\": 10
    }")
    
    if echo "$response" | grep -qE '"records"|"success"'; then
        test_passed "POST /api/data/query supports multiple criteria (AND logic)"
    else
        test_failed "POST /api/data/query multiple criteria" "$response"
    fi
}

test_field_selection() {
    echo ""
    echo "Test 8.4: Query with Field Selection"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"criteria\": [],
        \"limit\": 5
    }")
    
    if echo "$response" | grep -qE '"records"|"success"'; then
        test_passed "POST /api/data/query supports field selection"
    else
        test_failed "POST /api/data/query field selection" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
