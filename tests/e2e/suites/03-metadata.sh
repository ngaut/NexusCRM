#!/bin/bash
# tests/e2e/suites/03-metadata.sh
# Metadata Discovery Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Metadata Discovery"

run_suite() {
    section_header "$SUITE_NAME"
    
    test_get_all_apps
    test_get_all_schemas
    test_get_specific_schema
    test_nonexistent_schema
}

test_get_all_apps() {
    echo "Test 3.1: Get All Apps"
    
    local response=$(api_get "/api/metadata/apps")
    if echo "$response" | grep -q '"apps"' && echo "$response" | grep -q '\['; then
        local app_count=$(echo "$response" | grep -o '"id"' | wc -l)
        test_passed "GET /api/metadata/apps returns app list ($app_count apps)"
    else
        test_failed "GET /api/metadata/apps" "$response"
    fi
}

test_get_all_schemas() {
    echo ""
    echo "Test 3.2: Get All Schemas"
    
    local response=$(api_get "/api/metadata/schemas")
    if echo "$response" | grep -q '"schemas"' && echo "$response" | grep -q '\['; then
        local schema_count=$(echo "$response" | grep -o '"api_name"' | wc -l)
        test_passed "GET /api/metadata/schemas returns schema list ($schema_count objects)"
    else
        test_failed "GET /api/metadata/schemas" "$response"
    fi
}

test_get_specific_schema() {
    echo ""
    echo "Test 3.3: Get Specific Schema (User)"
    
    local response=$(api_get "/api/metadata/schemas/_System_User")
    if echo "$response" | grep -q '"schema"' && echo "$response" | grep -q '"_System_User"'; then
        local field_count=$(echo "$response" | grep -o '"api_name"' | wc -l)
        test_passed "GET /api/metadata/schemas/_System_User returns User schema ($field_count fields)"
    else
        test_failed "GET /api/metadata/schemas/_System_User" "$response"
    fi
}

test_nonexistent_schema() {
    echo ""
    echo "Test 3.4: Get Non-Existent Schema (404)"
    
    local response=$(api_get "/api/metadata/schemas/NonExistentObject12345")
    if echo "$response" | grep -qE "not found|Not Found"; then
        test_passed "GET non-existent schema returns 404"
    else
        test_failed "GET non-existent schema" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
