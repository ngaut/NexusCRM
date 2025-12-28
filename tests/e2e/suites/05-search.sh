#!/bin/bash
# tests/e2e/suites/05-search.sh
# Search Functionality Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Search Functionality"

run_suite() {
    section_header "$SUITE_NAME"
    
    test_global_search
    test_search_validation
}

test_global_search() {
    echo "Test 5.1: Global Search"
    
    local response=$(api_post "/api/data/search" '{"term": "Test"}')
    if echo "$response" | grep -q '"results"'; then
        test_passed "POST /api/data/search performs global search"
    else
        test_failed "POST /api/data/search" "$response"
    fi
}

test_search_validation() {
    echo ""
    echo "Test 5.2: Search Validation"
    
    local response=$(api_post "/api/data/search" '{}')
    if echo "$response" | grep -qE "required|term"; then
        test_passed "POST /api/data/search validates required term"
    else
        test_failed "POST /api/data/search validation" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
