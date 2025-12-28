#!/bin/bash
# tests/e2e/suites/10-master-detail.sh
# Master-Detail Relationship Full Lifecycle Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Master-Detail Relationship Tests"

# Helper for unique names
TS=$(date +%s)
PARENT_OBJ="parente2e_$TS"
CHILD_OBJ="childe2e_$TS"
TEST_FIELD="md_link"

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_create_objects
    test_create_md_field
    test_create_records
    
    # Optional cleanup
    # cleanup_objects
}

test_create_objects() {
    echo "Test 10.1: Create Parent and Child Objects"

    # Create Parent
    res_p=$(api_post "/api/metadata/schemas" "{
        \"label\": \"$PARENT_OBJ\",
        \"plural_label\": \"${PARENT_OBJ}s\",
        \"api_name\": \"$PARENT_OBJ\",
        \"is_custom\": true
    }")
    
    # Create Child
    res_c=$(api_post "/api/metadata/schemas" "{
        \"label\": \"$CHILD_OBJ\",
        \"plural_label\": \"${CHILD_OBJ}s\",
        \"api_name\": \"$CHILD_OBJ\",
        \"is_custom\": true
    }")

    if echo "$res_p" | grep -q "\"api_name\":\"$PARENT_OBJ\"" && echo "$res_c" | grep -q "\"api_name\":\"$CHILD_OBJ\""; then
        test_passed "Objects created: $PARENT_OBJ and $CHILD_OBJ"
    else
        test_failed "Object creation failed" "$res_p $res_c"
        return 1
    fi
}

test_create_md_field() {
    echo "Test 10.2: Create Master-Detail Field"

    local payload="{
        \"api_name\": \"$TEST_FIELD\",
        \"label\": \"Master Link\",
        \"type\": \"Lookup\",
        \"reference_to\": [\"$PARENT_OBJ\"],
        \"is_master_detail\": true,
        \"delete_rule\": \"CASCADE\",
        \"required\": true
    }"

    local response=$(api_post "/api/metadata/schemas/$CHILD_OBJ/fields" "$payload")

    if echo "$response" | grep -q "\"api_name\":\"$TEST_FIELD\""; then
        test_passed "Master-Detail Field created successfully"
    else
        test_failed "Field creation failed" "$response"
        return 1
    fi
    
    # Wait for permissions (simulated)
    # The refactor ensures this is fast, but cache invalidation might take a ms
    sleep 1
}

test_create_records() {
    echo "Test 10.3: Create Parent and Linked Child Records"

    # 1. Create Parent
    local p_payload="{\"name\": \"E2E Parent Record\"}"
    local p_res=$(api_post "/api/data/$PARENT_OBJ" "$p_payload")
    local p_id=$(json_extract "$p_res" "id")

    if [ -z "$p_id" ]; then
        test_failed "Parent record creation failed" "$p_res"
        return 1
    fi
    test_passed "Parent Record Created: $p_id"

    # 2. Create Child
    local c_payload="{\"name\": \"E2E Child Record\", \"$TEST_FIELD\": \"$p_id\"}"
    local c_res=$(api_post "/api/data/$CHILD_OBJ" "$c_payload")
    local c_id=$(json_extract "$c_res" "id")
    local link_val=$(json_extract "$c_res" "$TEST_FIELD")

    if [ -n "$c_id" ] && [ "$link_val" = "$p_id" ]; then
        test_passed "Child Record Created and Linked Correctly ($link_val)"
    else
        test_failed "Child record creation failed or link missing" "$c_res"
        return 1
    fi
}

cleanup_objects() {
    echo "Cleaning up..."
    api_delete "/api/metadata/schemas/$CHILD_OBJ" > /dev/null
    api_delete "/api/metadata/schemas/$PARENT_OBJ" > /dev/null
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
