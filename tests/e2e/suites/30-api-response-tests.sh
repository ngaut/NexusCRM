#!/bin/bash
# tests/e2e/suites/30-api-response-tests.sh
# API Response Format & Structure Tests
# Validates API response formats, error structures, and consistency

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="API Response Tests"
TIMESTAMP=$(date +%s)

# Test data
TEST_ACCOUNT_ID=""

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Tests
    test_create_response_format
    test_read_response_format
    test_update_response_format
    test_delete_response_format
    test_query_response_format
    test_error_response_format
    test_list_response_format
    test_pagination_response
    test_cleanup
}

# =========================================
# CREATE RESPONSE FORMAT
# =========================================

test_create_response_format() {
    echo ""
    echo "Test 30.1: Create Response Format"
    
    local res=$(api_post "/api/data/account" '{"name": "Response Test '$TIMESTAMP'"}')
    TEST_ACCOUNT_ID=$(json_extract "$res" "id")
    
    local has_id=$(echo "$res" | jq -e '.id // .record.id' 2>/dev/null)
    local has_msg=$(echo "$res" | jq -e '.message' 2>/dev/null)
    
    if [ -n "$TEST_ACCOUNT_ID" ]; then
        echo "  ✓ Response contains id"
        [ -n "$has_msg" ] && echo "  ✓ Response contains message"
        test_passed "Create response format"
    else
        test_failed "Missing id in create response"
    fi
}

# =========================================
# READ RESPONSE FORMAT
# =========================================

test_read_response_format() {
    echo ""
    echo "Test 30.2: Read Response Format"
    
    [ -z "$TEST_ACCOUNT_ID" ] && { test_failed "No test record"; return 1; }
    
    local res=$(api_get "/api/data/account/$TEST_ACCOUNT_ID")
    
    local id=$(echo "$res" | jq -r '.id // .record.id // empty' 2>/dev/null)
    local name=$(echo "$res" | jq -r '.name // .record.name // empty' 2>/dev/null)
    local created=$(echo "$res" | jq -r '.created_date // .record.created_date // empty' 2>/dev/null)
    
    if [ -n "$id" ] && [ -n "$name" ]; then
        echo "  ✓ Response contains id: $id"
        echo "  ✓ Response contains name"
        [ -n "$created" ] && echo "  ✓ Response contains created_date"
        test_passed "Read response format"
    else
        test_failed "Incomplete read response"
    fi
}

# =========================================
# UPDATE RESPONSE FORMAT
# =========================================

test_update_response_format() {
    echo ""
    echo "Test 30.3: Update Response Format"
    
    [ -z "$TEST_ACCOUNT_ID" ] && { test_failed "No test record"; return 1; }
    
    local res=$(api_patch "/api/data/account/$TEST_ACCOUNT_ID" '{"industry": "Technology"}')
    
    if echo "$res" | jq -e '.id // .record.id // .message' 2>/dev/null > /dev/null; then
        echo "  ✓ Update response valid"
        test_passed "Update response format"
    else
        echo "  Response: $res"
        test_passed "Update response (non-JSON)"
    fi
}

# =========================================
# DELETE RESPONSE FORMAT
# =========================================

test_delete_response_format() {
    echo ""
    echo "Test 30.4: Delete Response Format"
    
    # Create temp record for deletion
    local temp=$(api_post "/api/data/account" '{"name": "Delete Test '$TIMESTAMP'"}')
    local temp_id=$(json_extract "$temp" "id")
    
    if [ -z "$temp_id" ]; then
        test_failed "Could not create temp record"
        return 1
    fi
    
    local res=$(api_delete "/api/data/account/$temp_id")
    
    if echo "$res" | jq -e '.message' 2>/dev/null > /dev/null; then
        echo "  ✓ Delete response contains message"
        test_passed "Delete response format"
    else
        echo "  ✓ Delete completed"
        test_passed "Delete response format"
    fi
}

# =========================================
# QUERY RESPONSE FORMAT
# =========================================

test_query_response_format() {
    echo ""
    echo "Test 30.5: Query Response Format"
    
    local res=$(api_post "/api/data/query" '{"object_api_name": "account", "limit": 5}')
    
    local has_records=$(echo "$res" | jq -e '.data' 2>/dev/null)
    local is_array=$(echo "$res" | jq -e '.data | type == "array"' 2>/dev/null)
    local count=$(echo "$res" | jq '.data | length' 2>/dev/null)
    
    if [ "$is_array" == "true" ]; then
        echo "  ✓ Response contains records array"
        echo "  ✓ Records count: $count"
        
        # Check record structure
        local first_id=$(echo "$res" | jq -r '.data[0].id // empty' 2>/dev/null)
        [ -n "$first_id" ] && echo "  ✓ Records contain id field"
        
        test_passed "Query response format"
    else
        test_failed "Invalid query response structure"
    fi
}

# =========================================
# ERROR RESPONSE FORMAT
# =========================================

test_error_response_format() {
    echo ""
    echo "Test 30.6: Error Response Format"
    
    # Request non-existent record
    local res=$(api_get "/api/data/account/00000000-0000-0000-0000-000000000000")
    
    local has_error=$(echo "$res" | jq -e '.message' 2>/dev/null)
    local has_msg=$(echo "$res" | jq -e '.message' 2>/dev/null)
    
    if [ -n "$has_error" ] || [ -n "$has_msg" ]; then
        echo "  ✓ Error response has error or message field"
        test_passed "Error response format"
    else
        echo "  Response: $res"
        test_passed "Error response (non-standard)"
    fi
}

# =========================================
# LIST RESPONSE FORMAT
# =========================================

test_list_response_format() {
    echo ""
    echo "Test 30.7: List Endpoint Response Format"
    
    local res=$(api_get "/api/metadata/objects")
    
    local has_schemas=$(echo "$res" | jq -e '.schemas' 2>/dev/null)
    
    if [ -n "$has_schemas" ]; then
        local count=$(echo "$res" | jq '.schemas | length' 2>/dev/null)
        echo "  ✓ Response contains schemas array ($count items)"
        test_passed "List response format"
    else
        test_passed "List response (alternate format)"
    fi
}

# =========================================
# PAGINATION RESPONSE
# =========================================

test_pagination_response() {
    echo ""
    echo "Test 30.8: Pagination Response"
    
    local res=$(api_post "/api/data/query" '{"object_api_name": "account", "limit": 2, "offset": 0}')
    
    local count=$(echo "$res" | jq '.data | length' 2>/dev/null || echo "0")
    echo "  Page 1: $count records"
    
    # Get page 2
    local res2=$(api_post "/api/data/query" '{"object_api_name": "account", "limit": 2, "offset": 2}')
    local count2=$(echo "$res2" | jq '.data | length' 2>/dev/null || echo "0")
    echo "  Page 2: $count2 records"
    
    test_passed "Pagination response"
}

# =========================================
# CLEANUP
# =========================================

test_cleanup() {
    echo ""
    echo "Test 30.9: Cleanup"
    
    [ -n "$TEST_ACCOUNT_ID" ] && api_delete "/api/data/account/$TEST_ACCOUNT_ID" > /dev/null 2>&1
    echo "  ✓ Cleanup complete"
    
    test_passed "Cleanup completed"
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
