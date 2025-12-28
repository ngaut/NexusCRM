#!/bin/bash
# tests/e2e/suites/09-error-handling.sh
# Error Handling & Edge Cases Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Error Handling & Edge Cases"

run_suite() {
    section_header "$SUITE_NAME"
    
    test_malformed_json
    test_invalid_token
    test_missing_auth_header
    test_nonexistent_endpoint
    test_record_not_found
    test_invalid_object_type
}

test_malformed_json() {
    echo "Test 9.1: Malformed JSON Request"
    
    local response=$(api_post "/api/data/query" '{invalid json}')
    
    if echo "$response" | grep -qE "Bad Request|invalid|JSON|parse"; then
        test_passed "API properly handles malformed JSON"
    else
        test_failed "Malformed JSON handling" "$response"
    fi
}

test_invalid_token() {
    echo ""
    echo "Test 9.2: Invalid Token"
    
    local response=$(curl -s --max-time ${TIMEOUT:-30} "$BASE_URL/api/metadata/apps" \
        -H "Authorization: Bearer invalid_token_12345")
    
    if echo "$response" | grep -qE "Unauthorized|Invalid token|401"; then
        test_passed "API rejects invalid tokens"
    else
        test_failed "Invalid token handling" "$response"
    fi
}

test_missing_auth_header() {
    echo ""
    echo "Test 9.3: Missing Authorization Header"
    
    local response=$(api_get_unauth "/api/metadata/apps")
    if echo "$response" | grep -qE "Unauthorized|401|authorization"; then
        test_passed "API requires authorization header"
    else
        test_failed "Missing auth header handling" "$response"
    fi
}

test_nonexistent_endpoint() {
    echo ""
    echo "Test 9.4: Non-Existent Endpoint"
    
    local status=$(get_status_code "GET" "/api/nonexistent/endpoint")
    
    if [ "$status" = "404" ]; then
        test_passed "API returns 404 for non-existent endpoints"
    else
        test_failed "Non-existent endpoint" "HTTP $status (expected 404)"
    fi
}

test_record_not_found() {
    echo ""
    echo "Test 9.5: Record Not Found (404)"
    
    local response=$(api_get "/api/data/Account/00000000-0000-0000-0000-000000000000")
    if echo "$response" | grep -qE "not found|Not Found|404"; then
        test_passed "API returns 404 for non-existent records"
    else
        test_failed "Record not found handling" "$response"
    fi
}

test_invalid_object_type() {
    echo ""
    echo "Test 9.6: Invalid Object Type in Query"
    
    local response=$(api_post "/api/data/query" '{
        "objectApiName": "NonExistentObject12345",
        "criteria": [],
        "limit": 10
    }')
    
    if echo "$response" | grep -qE "unknown|not found|invalid|object"; then
        test_passed "Query rejects invalid object types"
    else
        test_failed "Invalid object type in query" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
