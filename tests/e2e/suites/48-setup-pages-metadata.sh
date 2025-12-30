#!/bin/bash
# tests/e2e/suites/48-setup-pages-metadata.sh
# Setup Pages Metadata Tests
# Verifies that setup page metadata is correctly loaded and accessible

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Setup Pages Metadata"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Re-authenticate for fresh token
    echo "Re-authenticating for tests..."
    api_login "$TEST_EMAIL" "$TEST_PASSWORD"
    echo ""
    
    test_query_setup_pages
    test_automation_studio_exists
    test_all_core_pages_present
    test_filter_enabled_pages
}

test_query_setup_pages() {
    echo "Test 48.1: Query _System_SetupPage Table"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "_System_SetupPage"
    }')
    
    if echo "$response" | grep -q '"records"'; then
        local count=$(echo "$response" | grep -o '"id":' | wc -l)
        echo "  Found $count setup pages"
        test_passed "Query _System_SetupPage returns records"
    else
        test_failed "Query _System_SetupPage" "$response"
    fi
}

test_automation_studio_exists() {
    echo ""
    echo "Test 48.2: Automation Studio Setup Page Exists"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "_System_SetupPage",
        "filter_expr": "component_name == '\''Flows'\''"
    }')
    
    if echo "$response" | grep -q '"Automation Studio"'; then
        test_passed "Automation Studio setup page exists with correct component"
    elif echo "$response" | grep -q '"component_name":"Flows"'; then
        test_passed "Automation Studio setup page exists (Flows component)"
    else
        test_failed "Automation Studio setup page not found" "$response"
    fi
}

test_all_core_pages_present() {
    echo ""
    echo "Test 48.3: All Core Setup Pages Present"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "_System_SetupPage",
        "filter_expr": "is_enabled == true"
    }')
    
    local missing=""
    
    # Check for required pages
    for page in "Users" "Profiles" "Object Manager" "App Manager" "Automation Studio" "Data Import" "Data Export" "Recycle Bin"; do
        if ! echo "$response" | grep -q "\"$page\""; then
            missing="$missing $page"
        fi
    done
    
    if [ -z "$missing" ]; then
        test_passed "All 8 core setup pages are present"
    else
        test_failed "Missing setup pages:$missing"
    fi
}

test_filter_enabled_pages() {
    echo ""
    echo "Test 48.4: Filter Expression with is_enabled Field"
    
    local response=$(api_post "/api/data/query" '{
        "object_api_name": "_System_SetupPage",
        "filter_expr": "is_enabled == true"
    }')
    
    if echo "$response" | grep -q '"records"'; then
        local count=$(echo "$response" | grep -o '"is_enabled":true' | wc -l)
        echo "  Found $count enabled setup pages"
        if [ $count -ge 8 ]; then
            test_passed "Filter by is_enabled returns expected results"
        else
            test_failed "Expected at least 8 enabled pages, got $count"
        fi
    else
        test_failed "Filter by is_enabled" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
