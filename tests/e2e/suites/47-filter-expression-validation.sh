#!/bin/bash
# tests/e2e/suites/47-filter-expression-validation.sh
# Filter Expression Validation Tests
# Tests that filter expressions correctly handle operators and return errors for invalid syntax

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Filter Expression Validation"

# Helper for unique names
TS=$(date +%s)
TEST_OBJ="filter_expr_test_$TS"

test_cleanup() {
    echo "Cleaning up test object..."
    api_delete "/api/metadata/objects/$TEST_OBJ" > /dev/null 2>&1
}
trap test_cleanup EXIT

run_suite() {
    section_header "$SUITE_NAME"
    
    # Re-authenticate for fresh token
    echo "Re-authenticating for tests..."
    api_login "$TEST_EMAIL" "$TEST_PASSWORD"
    echo ""
    
    setup_test_object
    seed_test_data
    
    test_valid_and_operator
    test_invalid_and_operator
    test_valid_or_operator
    test_invalid_or_operator
    test_combined_operators
}

setup_test_object() {
    echo "Setup: Creating test object '$TEST_OBJ'..."
    
    local response=$(api_post "/api/metadata/objects" "{
        \"label\": \"$TEST_OBJ\",
        \"plural_label\": \"${TEST_OBJ}s\",
        \"api_name\": \"$TEST_OBJ\",
        \"description\": \"E2E Filter Expression Validation Test Object\",
        \"is_custom\": true
    }")
    
    if echo "$response" | grep -q "\"api_name\":\"$TEST_OBJ\""; then
        echo "  ✓ Test object created: $TEST_OBJ"
    else
        echo "  ✗ Failed to create test object"
        echo "  Response: $response"
        return 1
    fi
    
    # Add fields
    api_post "/api/metadata/objects/$TEST_OBJ/fields" '{"api_name": "status", "label": "Status", "type": "Text"}' > /dev/null
    api_post "/api/metadata/objects/$TEST_OBJ/fields" '{"api_name": "is_read", "label": "Is Read", "type": "Checkbox"}' > /dev/null
    echo "  ✓ Fields added to test object"
    
    # Wait for Schema Cache (Polling)
    echo "  Waiting for field 'status'..."
    for i in {1..10}; do
        meta=$(api_get "/api/metadata/objects/$TEST_OBJ")
        if echo "$meta" | grep -q "\"api_name\":\"status\""; then
            break
        fi
        sleep 0.5
    done
}

seed_test_data() {
    echo "Setup: Seeding test data..."
    api_post "/api/data/$TEST_OBJ" '{"name": "Test Record 1", "status": "Open", "is_read": false}' > /dev/null
    api_post "/api/data/$TEST_OBJ" '{"name": "Test Record 2", "status": "Closed", "is_read": true}' > /dev/null
    echo "  ✓ 2 test records seeded"
}

test_valid_and_operator() {
    echo ""
    echo "Test 47.1: Valid && Operator"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"filter_expr\": \"status == 'Open' && is_read == false\"
    }")
    
    if echo "$response" | grep -q '"records"'; then
        test_passed "Valid && operator works correctly"
    else
        test_failed "Valid && operator" "$response"
    fi
}

test_invalid_and_operator() {
    echo ""
    echo "Test 47.2: Invalid AND Operator"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"filter_expr\": \"status == 'Open' AND is_read == false\"
    }")
    
    if echo "$response" | grep -qE '"error"|"Error"'; then
        test_passed "Invalid AND operator correctly rejected"
    else
        # If it returns records, that means AND was incorrectly interpreted
        if echo "$response" | grep -q '"records"'; then
            test_failed "Invalid AND operator was incorrectly accepted" "$response"
        else
            test_passed "Invalid AND operator correctly rejected"
        fi
    fi
}

test_valid_or_operator() {
    echo ""
    echo "Test 47.3: Valid || Operator"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"filter_expr\": \"status == 'Open' || status == 'Closed'\"
    }")
    
    if echo "$response" | grep -q '"records"'; then
        test_passed "Valid || operator works correctly"
    else
        test_failed "Valid || operator" "$response"
    fi
}

test_invalid_or_operator() {
    echo ""
    echo "Test 47.4: Invalid OR Operator"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"filter_expr\": \"status == 'Open' OR status == 'Closed'\"
    }")
    
    if echo "$response" | grep -qE '"error"|"Error"'; then
        test_passed "Invalid OR operator correctly rejected"
    else
        if echo "$response" | grep -q '"records"'; then
            test_failed "Invalid OR operator was incorrectly accepted" "$response"
        else
            test_passed "Invalid OR operator correctly rejected"
        fi
    fi
}

test_combined_operators() {
    echo ""
    echo "Test 47.5: Combined Valid Operators"
    
    local response=$(api_post "/api/data/query" "{
        \"object_api_name\": \"$TEST_OBJ\",
        \"filter_expr\": \"(status == 'Open' || status == 'Closed') && name != null\"
    }")
    
    if echo "$response" | grep -q '"records"'; then
        test_passed "Combined valid operators work correctly"
    else
        test_failed "Combined valid operators" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
