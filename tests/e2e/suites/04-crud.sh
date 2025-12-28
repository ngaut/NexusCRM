#!/bin/bash
# tests/e2e/suites/04-crud.sh
# Record CRUD Operations Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Record CRUD Operations"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate first
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_create_record
    test_query_records
    test_query_criteria
    test_get_record
    test_update_record
    test_delete_record
    test_delete_record
    test_lead_lifecycle
    test_frontend_simulation
}

# Test 4.1: Create Record (Account)
test_create_record() {
    echo "Test 4.1: Create Record (Account)"
    
    # Create account
    response=$(api_post "/api/data/account" '{"name": "E2E Test Account", "industry": "Technology", "annual_revenue": 5000000}')
    
    # Extract ID
    record_id=$(json_extract "$response" "id")
    
    if [ -n "$record_id" ]; then
        test_passed "POST /api/data/account creates record (ID: $record_id)"
        echo "$record_id" > /tmp/e2e_record_id
    else
        test_failed "POST /api/data/account failed" "$response"
        return 1
    fi
}

# Test 4.2: Query Records
test_query_records() {
    echo "Test 4.2: Query Records"
    
    response=$(api_post "/api/data/query" '{"object_api_name": "account", "limit": 10}')
    
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
    
    response=$(api_post "/api/data/query" '{"object_api_name": "account", "criteria": [{"field": "industry", "op": "=", "val": "Technology"}]}')
    
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
    response=$(api_get "/api/data/account/$record_id")
    
    name=$(json_extract "$response" "name")
    
    if [ "$name" = "E2E Test Account" ]; then
        test_passed "GET /api/data/account/$record_id returns correct record"
    else
        test_failed "GET /api/data/account/$record_id failed" "$response"
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
    response=$(api_patch "/api/data/account/$record_id" '{"name": "E2E Test Account (Updated)", "annual_revenue": 7500000}')
    
    # Verify update
    response=$(api_get "/api/data/account/$record_id")
    name=$(json_extract "$response" "name")
    
    # revenue check via grep
    if [ "$name" = "E2E Test Account (Updated)" ] && echo "$response" | grep -q "7500000"; then
        test_passed "PATCH /api/data/account/$record_id updates record correctly"
    else
        test_failed "PATCH /api/data/account/$record_id failed verification" "$response"
    fi
}

test_delete_record() {
    echo "Test 4.6: Delete Record (Soft Delete)"

    if [ ! -f /tmp/e2e_record_id ]; then
        echo "Skipping Test 4.6 - No record ID"
        return
    fi
    
    record_id=$(cat /tmp/e2e_record_id)
    
    status=$(get_status_code "DELETE" "/api/data/account/$record_id")
    if [ "$status" = "200" ] || [ "$status" = "204" ]; then
        test_passed "DELETE /api/data/account/:id soft deletes record"
    else
        test_failed "DELETE /api/data/account/$record_id failed (HTTP $status)" ""
    fi
}

test_lead_lifecycle() {
    echo "Test 4.7: Lead Lifecycle & Validation"
    
    # 1. Validation Fail (Missing Email)
    echo "  Testing validation failure (missing email)..."
    local fail_response=$(api_post "/api/data/lead" '{"name": "No Email Lead", "company": "Fail Co", "status": "New"}')
    if echo "$fail_response" | grep -q 'required' || echo "$fail_response" | grep -q "validation"; then
        test_passed "Deep validation caught missing Email"
    else
        if echo "$fail_response" | grep -q "\"id\""; then
             test_failed "Backend accepted Lead without Email (Expected Validation Error)" "$fail_response"
        else
             test_passed "Creation failed as expected"
        fi
    fi

    # 2. Success Create
    echo "  Testing success creation..."
    local success_response=$(api_post "/api/data/lead" '{"name": "Valid Lead", "company": "Success Co", "email": "valid@test.com", "status": "New"}')
    local lead_id=$(json_extract "$success_response" "id")
    
    if [ -n "$lead_id" ] && [ "$lead_id" != "null" ]; then
        test_passed "Created Valid Lead (ID: $lead_id)"
        # Cleanup
        api_delete "/api/data/lead/$lead_id" > /dev/null
    else
        test_failed "Failed to create valid lead" "$success_response"
    fi
}


test_frontend_simulation() {
    echo "Test 4.7: Frontend Simulation (Object Creation Flow)"
    
    local object_name="UiTestObject"
    local object_api="ui_test_object"

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
