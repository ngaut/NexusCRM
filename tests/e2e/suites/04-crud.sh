#!/bin/bash
# tests/e2e/suites/04-crud.sh
# Record CRUD Operations Tests
# REFACTORED: Uses dynamically created test objects instead of hardcoded Account/Lead

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Record CRUD Operations"

# Helper for unique names
TS=$(date +%s)
TEST_OBJ="crudobject_$TS"

test_cleanup() {
    echo "Cleaning up test object..."
    api_delete "/api/metadata/schemas/$TEST_OBJ" > /dev/null 2>&1
}
trap test_cleanup EXIT

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate first
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    # Create test object for this suite
    setup_test_object
    
    test_create_record
    test_query_records
    test_query_criteria
    test_get_record
    test_update_record
    test_delete_record
    test_frontend_simulation
}

setup_test_object() {
    echo "Setup: Creating test object '$TEST_OBJ'..."
    
    local response=$(api_post "/api/metadata/schemas" "{
        \"label\": \"$TEST_OBJ\",
        \"plural_label\": \"${TEST_OBJ}s\",
        \"api_name\": \"$TEST_OBJ\",
        \"description\": \"E2E Test Object\",
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
    api_post "/api/metadata/schemas/$TEST_OBJ/fields" '{"api_name": "industry", "label": "Industry", "type": "Text"}' > /dev/null
    api_post "/api/metadata/schemas/$TEST_OBJ/fields" '{"api_name": "annual_revenue", "label": "Annual Revenue", "type": "Number"}' > /dev/null
    echo "  ✓ Fields added to test object"
    
    sleep 1  # Allow caches to refresh
}

# Test 4.1: Create Record
test_create_record() {
    echo "Test 4.1: Create Record ($TEST_OBJ)"
    
    response=$(api_post "/api/data/$TEST_OBJ" '{"name": "E2E Test Record", "industry": "Technology", "annual_revenue": 5000000}')
    
    # Extract ID
    record_id=$(json_extract "$response" "id")
    
    if [ -n "$record_id" ]; then
        test_passed "POST /api/data/$TEST_OBJ creates record (ID: $record_id)"
        echo "$record_id" > /tmp/e2e_record_id
    else
        test_failed "POST /api/data/$TEST_OBJ failed" "$response"
        return 1
    fi
}

# Test 4.2: Query Records
test_query_records() {
    echo "Test 4.2: Query Records"
    
    response=$(api_post "/api/data/query" "{\"object_api_name\": \"$TEST_OBJ\", \"limit\": 10}")
    
    if echo "$response" | grep -q '"id"'; then
        count=$(echo "$response" | grep -o '"id":' | wc -l)
        test_passed "POST /api/data/query returns records ($count found)"
    else
        test_failed "POST /api/data/query returned no records" "$response"
    fi
}

# Test 4.3: Query with Criteria
test_query_criteria() {
    echo "Test 4.3: Query with Criteria"
    
    response=$(api_post "/api/data/query" "{\"object_api_name\": \"$TEST_OBJ\", \"criteria\": [{\"field\": \"industry\", \"op\": \"=\", \"val\": \"Technology\"}]}")
    
    if echo "$response" | grep -q '"id"'; then
        test_passed "POST /api/data/query with criteria returns matching records"
    else
        test_failed "POST /api/data/query with criteria failed" "$response"
    fi
}

# Test 4.4: Get Specific Record by ID
test_get_record() {
    echo "Test 4.4: Get Specific Record by ID"
    
    if [ ! -f /tmp/e2e_record_id ]; then
        echo "Skipping test - no record ID from previous step"
        return 1
    fi
    
    record_id=$(cat /tmp/e2e_record_id)
    response=$(api_get "/api/data/$TEST_OBJ/$record_id")
    
    name=$(json_extract "$response" "name")
    
    if [ "$name" = "E2E Test Record" ]; then
        test_passed "GET /api/data/$TEST_OBJ/$record_id returns correct record"
    else
        test_failed "GET /api/data/$TEST_OBJ/$record_id failed" "$response"
    fi
}

# Test 4.5: Update Record
test_update_record() {
    echo "Test 4.5: Update Record"
    
    if [ ! -f /tmp/e2e_record_id ]; then
        echo "Skipping test - no record ID from previous step"
        return 1
    fi
    
    record_id=$(cat /tmp/e2e_record_id)
    response=$(api_patch "/api/data/$TEST_OBJ/$record_id" '{"name": "E2E Test Record (Updated)", "annual_revenue": 7500000}')
    
    # Verify update
    response=$(api_get "/api/data/$TEST_OBJ/$record_id")
    name=$(json_extract "$response" "name")
    
    if [ "$name" = "E2E Test Record (Updated)" ] && echo "$response" | grep -q "7500000"; then
        test_passed "PATCH /api/data/$TEST_OBJ/$record_id updates record correctly"
    else
        test_failed "PATCH /api/data/$TEST_OBJ/$record_id failed verification" "$response"
    fi
}

test_delete_record() {
    echo "Test 4.6: Delete Record (Soft Delete)"

    if [ ! -f /tmp/e2e_record_id ]; then
        echo "Skipping Test 4.6 - No record ID"
        return
    fi
    
    record_id=$(cat /tmp/e2e_record_id)
    
    status=$(get_status_code "DELETE" "/api/data/$TEST_OBJ/$record_id")
    if [ "$status" = "200" ] || [ "$status" = "204" ]; then
        test_passed "DELETE /api/data/$TEST_OBJ/:id soft deletes record"
    else
        test_failed "DELETE /api/data/$TEST_OBJ/$record_id failed (HTTP $status)" ""
    fi
}

test_frontend_simulation() {
    echo "Test 4.7: Frontend Simulation (Object Creation Flow)"
    
    local object_name="UiTestObject"
    local object_api="ui_test_object_$TS"

    # 1. Create Object (POST /api/metadata/schemas)
    echo "Step 1: Creating object via API (simulating frontend)..."
    local response=$(api_post "/api/metadata/schemas" "{
        \"label\": \"$object_name\",
        \"plural_label\": \"${object_name}s\",
        \"api_name\": \"$object_api\",
        \"description\": \"Created via E2E Simulation\",
        \"is_custom\": true,
        \"searchable\": true
    }")

    if echo "$response" | grep -q "\"api_name\":\"$object_api\""; then
        test_passed "Frontend Object Creation (POST /api/metadata/schemas)"
    else
        test_failed "Frontend Object Creation" "$response"
        return
    fi

    # 2. Verify Object Existence (GET /api/metadata/schemas/:apiName)
    echo "Step 2: Verifying object existence (simulating ObjectDetail load)..."
    local detail_response=$(api_get "/api/metadata/schemas/$object_api")

    if echo "$detail_response" | grep -q "\"api_name\":\"$object_api\""; then
        test_passed "Object Retrieval (GET /api/metadata/schemas/$object_api)"
    else
        test_failed "Object Retrieval" "$detail_response"
    fi
    
    # Cleanup
    echo "Step 3: Cleanup"
    api_delete "/api/metadata/schemas/$object_api" > /dev/null
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
