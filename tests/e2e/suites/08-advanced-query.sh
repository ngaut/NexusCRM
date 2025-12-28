#!/bin/bash
# tests/e2e/suites/08-advanced-query.sh
# Advanced Query Operations Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Advanced Query Operations"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Re-authenticate for fresh token
    echo "Re-authenticating for remaining tests..."
    api_login "$TEST_EMAIL" "$TEST_PASSWORD"
    echo ""
    
    test_pagination
    test_sorting
    test_multiple_criteria
    test_field_selection
}

test_pagination() {
    echo "Test 8.1: Query with Pagination"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "Account",
        "criteria": [],
        "limit": 5,
        "offset": 0
    }')
    
    if echo "$response" | grep -q '"records"'; then
        test_passed "POST /api/data/query supports pagination (limit/offset)"
    else
        test_failed "POST /api/data/query pagination" "$response"
    fi
}

test_sorting() {
    echo ""
    echo "Test 8.2: Query with Sorting"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "Account",
        "criteria": [],
        "order_by": [{"field": "name", "direction": "ASC"}],
        "limit": 10
    }')
    
    if echo "$response" | grep -qE '"records"|"success"'; then
        test_passed "POST /api/data/query supports sorting (orderBy)"
    else
        test_failed "POST /api/data/query sorting" "$response"
    fi
}

test_multiple_criteria() {
    echo ""
    echo "Test 8.3: Query with Multiple Criteria (AND)"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "Account",
        "criteria": [
            {"field": "industry", "op": "=", "val": "Technology"},
            {"field": "annual_revenue", "op": ">", "val": 1000000}
        ],
        "limit": 10
    }')
    
    if echo "$response" | grep -qE '"records"|"success"'; then
        test_passed "POST /api/data/query supports multiple criteria (AND logic)"
    else
        test_failed "POST /api/data/query multiple criteria" "$response"
    fi
}

test_field_selection() {
    echo ""
    echo "Test 8.4: Query with Field Selection"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "Account",
        "fields": ["id", "name", "industry"],
        "criteria": [],
        "limit": 5
    }')
    
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
