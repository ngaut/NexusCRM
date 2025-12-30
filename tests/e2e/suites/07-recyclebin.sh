#!/bin/bash
# tests/e2e/suites/07-recyclebin.sh
# Recycle Bin Operations Tests
# REFACTORED: Uses dynamically created test objects instead of hardcoded Lead

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Recycle Bin Operations"

# Helper for unique names
TS=$(date +%s)
TEST_OBJ="recycle_test_$TS"

test_cleanup() {
    echo "Cleaning up test object..."
    api_delete "/api/metadata/schemas/$TEST_OBJ" > /dev/null 2>&1
}
trap test_cleanup EXIT

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    setup_test_object
    test_recycle_bin_lifecycle
}

setup_test_object() {
    echo "Setup: Creating test object '$TEST_OBJ'..."
    
    local response=$(api_post "/api/metadata/schemas" "{
        \"label\": \"$TEST_OBJ\",
        \"plural_label\": \"${TEST_OBJ}s\",
        \"api_name\": \"$TEST_OBJ\",
        \"description\": \"E2E Recycle Bin Test Object\",
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
    
    sleep 1  # Allow caches to refresh
}

test_recycle_bin_lifecycle() {
    echo "Test 7.1: Full Recycle Bin Lifecycle (Create -> Delete -> Restore -> Purge)"
    
    # 1. Create a Record to be deleted
    echo "  Creating temporary record..."
    local create_payload="{\"name\": \"Recycle Bin Test Record\"}"
    local response=$(api_post "/api/data/$TEST_OBJ" "$create_payload")
    local record_id=$(json_extract "$response" "id")
    
    if [ -z "$record_id" ] || [ "$record_id" == "null" ]; then
        test_failed "Setup: Create record failed" "$response"
        return
    fi
    test_passed "Setup: Created record (ID: $record_id)"

    # 2. Soft Delete the Record
    echo "  Soft deleting record..."
    local status=$(get_status_code "DELETE" "/api/data/$TEST_OBJ/$record_id")
    if [ "$status" != "200" ] && [ "$status" != "204" ]; then
         test_failed "Soft Delete failed" "HTTP $status"
         return
    fi
    test_passed "Soft Deleted Record"

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
    response=$(api_get "/api/data/$TEST_OBJ/$record_id")
    local check_id=$(json_extract "$response" "id")
    if [ "$check_id" == "$record_id" ]; then
        test_passed "Restored record accessible via API"
    else
        test_failed "Restored record not accessible" "$response"
    fi

    # 6. Delete Again for Purge
    echo "  Deleting again for purge testing..."
    api_delete "/api/data/$TEST_OBJ/$record_id" > /dev/null

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
    status=$(get_status_code "GET" "/api/data/$TEST_OBJ/$record_id")
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
