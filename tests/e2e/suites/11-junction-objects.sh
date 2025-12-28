#!/bin/bash
# tests/e2e/suites/11-junction-objects.sh
# Junction Object (Many-to-Many) Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Junction Object (Many-to-Many) Tests"

# Helper for unique names
TS=$(date +%s)
PARENT_1="parenta_$TS"
PARENT_2="parentb_$TS"
PARENT_3="parentc_$TS"
JUNCTION_OBJ="junctionobj_$TS"

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_create_objects
    test_create_2_md_fields
    test_create_3rd_md_field_fail
    test_data_and_cascade
    
    # cleanup_objects
}

test_create_objects() {
    echo "Test 11.1: Create 3 Parents and 1 Junction Object"

    api_post "/api/metadata/schemas" "{\"api_name\": \"$PARENT_1\", \"label\": \"$PARENT_1\", \"is_custom\": true}" > /dev/null
    api_post "/api/metadata/schemas" "{\"api_name\": \"$PARENT_2\", \"label\": \"$PARENT_2\", \"is_custom\": true}" > /dev/null
    api_post "/api/metadata/schemas" "{\"api_name\": \"$PARENT_3\", \"label\": \"$PARENT_3\", \"is_custom\": true}" > /dev/null

    res_j=$(api_post "/api/metadata/schemas" "{\"api_name\": \"$JUNCTION_OBJ\", \"label\": \"$JUNCTION_OBJ\", \"is_custom\": true}")
    
    if echo "$res_j" | grep -q "\"api_name\":\"$JUNCTION_OBJ\""; then
        test_passed "Junction Object Created"
    else
        test_failed "Junction Object Creation Failed" "$res_j"
        return 1
    fi
}

test_create_2_md_fields() {
    echo "Test 11.2: Create 2 Master-Detail Fields (Allowed)"

    # MD 1 -> Parent 1
    local res1=$(api_post "/api/metadata/schemas/$JUNCTION_OBJ/fields" "{
        \"api_name\": \"md_link_1\", \"label\": \"Link 1\", \"type\": \"Lookup\",
        \"reference_to\": [\"$PARENT_1\"], \"is_master_detail\": true, \"required\": true
    }")

    # MD 2 -> Parent 2
    local res2=$(api_post "/api/metadata/schemas/$JUNCTION_OBJ/fields" "{
        \"api_name\": \"md_link_2\", \"label\": \"Link 2\", \"type\": \"Lookup\",
        \"reference_to\": [\"$PARENT_2\"], \"is_master_detail\": true, \"required\": true
    }")

    if echo "$res1" | grep -q "\"api_name\":\"md_link_1\"" && echo "$res2" | grep -q "\"api_name\":\"md_link_2\""; then
        test_passed "Successfully created 2 Master-Detail fields"
    else
        test_failed "Failed to create 2 MD fields" "$res1 $res2"
        return 1
    fi
}

test_create_3rd_md_field_fail() {
    echo "Test 11.3: Create 3rd Master-Detail Field (Should Fail)"

    # MD 3 -> Parent 3 (Should be blocked)
    local res3=$(api_post "/api/metadata/schemas/$JUNCTION_OBJ/fields" "{
        \"api_name\": \"md_link_3\", \"label\": \"Link 3\", \"type\": \"Lookup\",
        \"reference_to\": [\"$PARENT_3\"], \"is_master_detail\": true, \"required\": true
    }")

    # Expecting 400 Bad Request
    if echo "$res3" | grep -q "maximum of 2 Master-Detail relationships"; then
        test_passed "Correctly blocked 3rd Master-Detail field"
    else
        test_failed "Failed to block 3rd MD field (or wrong error)" "$res3"
        return 1
    fi
}

test_data_and_cascade() {
    echo "Test 11.4: Verify Cascade Delete from Parent 1"

    # 1. Create Parent 1
    local p1_res=$(api_post "/api/data/$PARENT_1" "{\"name\": \"Parent 1\"}")
    local p1_id=$(json_extract "$p1_res" "id")

    # 2. Create Parent 2
    local p2_res=$(api_post "/api/data/$PARENT_2" "{\"name\": \"Parent 2\"}")
    local p2_id=$(json_extract "$p2_res" "id")

    # 3. Create Junction Record
    local j_res=$(api_post "/api/data/$JUNCTION_OBJ" "{\"name\": \"Junction Item\", \"md_link_1\": \"$p1_id\", \"md_link_2\": \"$p2_id\"}")
    local j_id=$(json_extract "$j_res" "id")

    if [ -z "$j_id" ]; then
        test_failed "Failed to create junction record" "$j_res"
        return 1
    fi

    # 4. Delete Parent 1
    api_delete "/api/data/$PARENT_1/$p1_id" > /dev/null

    # 5. Verify Junction Deleted
    local status=$(get_status_code "GET" "/api/data/$JUNCTION_OBJ/$j_id")
    if [ "$status" = "404" ]; then
        test_passed "Cascaded delete confirmed (Junction record deleted after Parent 1 delete)"
    else
        # Check is_deleted flag if soft delete
        local check=$(api_get "/api/data/$JUNCTION_OBJ/$j_id")
        if echo "$check" | grep -q "\"is_deleted\":true" || echo "$check" | grep -q "\"is_deleted\":1"; then
             test_passed "Cascaded delete confirmed (Soft Deleted)"
        else 
             test_failed "Cascade delete failed. Status: $status" "$check"
        fi
    fi
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
