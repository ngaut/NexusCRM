#!/bin/bash
# tests/e2e/suites/07-recyclebin.sh
# Recycle Bin Operations Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Recycle Bin Operations"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_recycle_bin_lifecycle
}

test_recycle_bin_lifecycle() {
    echo "Test 7.1: Full Recycle Bin Lifecycle (Create -> Delete -> Restore -> Purge)"
    
    # 1. Create a Record to be deleted
    echo "  Creating temporary Lead..."
    local create_payload='{"name": "Recycle Bin Test", "company": "Recycle Co", "email": "recycle@test.co", "status": "New"}'
    local response=$(api_post "/api/data/lead" "$create_payload")
    local record_id=$(json_extract "$response" "id")
    
    if [ -z "$record_id" ] || [ "$record_id" == "null" ]; then
        test_failed "Setup: Create Lead failed" "$response"
        return
    fi
    test_passed "Setup: Created Lead (ID: $record_id)"

    # 2. Soft Delete the Record
    echo "  Soft deleting lead..."
    local status=$(get_status_code "DELETE" "/api/data/lead/$record_id")
    if [ "$status" != "200" ] && [ "$status" != "204" ]; then
         test_failed "Soft Delete failed" "HTTP $status"
         return
    fi
    test_passed "Soft Deleted Lead"

    # 3. Verify in Recycle Bin (My Scope)
    echo "  Verifying in Recycle Bin (Scope: Mine)..."
    response=$(api_get "/api/data/recyclebin/items?scope=mine")
    if echo "$response" | grep -q "$record_id"; then
        test_passed "Record found in 'My' Recycle Bin"
    else
        test_failed "Record NOT found in 'My' Recycle Bin" "$response"
    fi

    # 4. Verify in Recycle Bin (All Scope - Admin)
    # Note: We are logged in as Admin, so this should work
    echo "  Verifying in Recycle Bin (Scope: All)..."
    response=$(api_get "/api/data/recyclebin/items?scope=all")
    if echo "$response" | grep -q "$record_id"; then
        test_passed "Record found in 'Org' Recycle Bin"
    else
        test_failed "Record NOT found in 'Org' Recycle Bin" "$response"
    fi

    # 5. Restore Record
    echo "  Restoring record..."
    status=$(get_status_code "POST" "/api/data/recyclebin/restore/$record_id" '{}')
    
    if [ "$status" != "200" ]; then
        test_failed "Restore failed" "HTTP $status"
        return
    fi
    
    # Verify it's back in Data API
    response=$(api_get "/api/data/lead/$record_id")
    local check_id=$(json_extract "$response" "id")
    if [ "$check_id" == "$record_id" ]; then
        test_passed "Restored record accessible via API"
    else
        test_failed "Restored record not accessible" "$response"
    fi

    # 6. Delete Again for Purge
    echo "  Deleting again for purge testing..."
    api_delete "/api/data/lead/$record_id" > /dev/null

    # 7. Purge (Permanent Delete)
    echo "  Purging record..."
    status=$(get_status_code "DELETE" "/api/data/recyclebin/$record_id")
    if [ "$status" != "200" ] && [ "$status" != "204" ]; then
        test_failed "Purge failed" "HTTP $status"
        return
    fi
    test_passed "Purged record"

    # 8. Verify Gone
    echo "  Verifying permanent deletion..."
    # Should not be in Bin
    response=$(api_get "/api/data/recyclebin/items?scope=mine")
    if echo "$response" | grep -q "$record_id"; then
        test_failed "Record still in Recycle Bin after purge" ""
    else
        test_passed "Record removed from Recycle Bin"
    fi
    
    # Should not be in Data API (404)
    status=$(get_status_code "GET" "/api/data/lead/$record_id")
    if [ "$status" == "404" ]; then
        test_passed "Record 404s in Data API (Permanently Deleted)"
    else
        test_failed "Record still accessible in Data API" "HTTP $status"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
