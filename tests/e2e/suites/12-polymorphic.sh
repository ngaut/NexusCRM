#!/bin/bash
# tests/e2e/suites/12-polymorphic.sh
# Polymorphic Lookups E2E Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Polymorphic Lookups"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate first
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_polymorphic_flow
}

test_polymorphic_flow() {
    echo "Test 12.1: Polymorphic Reference Flow"
    
    # Generate unique names
    local ts=$(date +%s)
    local p1_name="polyacc_$ts"
    local p2_name="polyopp_$ts"
    local child_name="polyevent_$ts"
    
    # 1. Create Parent Objects
    echo "Creating parent objects..."
    api_post "/api/metadata/schemas" "{\"label\": \"$p1_name\", \"api_name\": \"$p1_name\", \"is_custom\": true}" > /dev/null
    api_post "/api/metadata/schemas" "{\"label\": \"$p2_name\", \"api_name\": \"$p2_name\", \"is_custom\": true}" > /dev/null
    
    # 2. Create Child Object
    echo "Creating child object..."
    api_post "/api/metadata/schemas" "{\"label\": \"$child_name\", \"api_name\": \"$child_name\", \"is_custom\": true}" > /dev/null
    
    # 3. Add Polymorphic Field
    echo "Adding polymorphic lookup field 'what_id'..."
    # reference_to is sent as JSON array
    local field_payload="{\"api_name\": \"what_id\", \"label\": \"Related To\", \"type\": \"Lookup\", \"reference_to\": [\"$p1_name\", \"$p2_name\"]}"
    local field_res=$(api_post "/api/metadata/schemas/$child_name/fields" "$field_payload")
    
    if echo "$field_res" | grep -q '"api_name":"what_id"'; then
        echo "  Field created successfully."
    else
        test_failed "Field Creation Failed" "$field_res"
        return 1
    fi
    
    # 4. Create Parent Records
    echo "Creating parent records..."
    local p1_res=$(api_post "/api/data/$p1_name" '{"name": "Parent 1"}')
    local p1_id=$(json_extract "$p1_res" "id")
    
    local p2_res=$(api_post "/api/data/$p2_name" '{"name": "Parent 2"}')
    local p2_id=$(json_extract "$p2_res" "id")
    
    if [ -z "$p1_id" ] || [ -z "$p2_id" ]; then
        test_failed "Parent Record Creation Failed" "$p1_res $p2_res"
        return 1
    fi
    
    # 5. Create Child Records linking to P1 and P2
    echo "Creating child records..."
    local c1_res=$(api_post "/api/data/$child_name" "{\"name\": \"Event 1\", \"what_id\": \"$p1_id\"}")
    local c1_id=$(json_extract "$c1_res" "id")
    
    local c2_res=$(api_post "/api/data/$child_name" "{\"name\": \"Event 2\", \"what_id\": \"$p2_id\"}")
    local c2_id=$(json_extract "$c2_res" "id")
    
    if [ -z "$c1_id" ] || [ -z "$c2_id" ]; then
        test_failed "Child Record Creation Failed" "$c1_res $c2_res"
        return 1
    fi
    api_get "/api/data/$child_name/$c1_id" > /dev/null # Ensure creation verified
    
    # 6. Verify Read Path (Check _type)
    echo "Verifying read path metadata..."
    local c1_get=$(api_get "/api/data/$child_name/$c1_id")
    local c2_get=$(api_get "/api/data/$child_name/$c2_id")
    
    # JSON extract of what_id_type
    local type1=$(echo "$c1_get" | grep -o "\"what_id_type\":\"[^\"]*\"" | cut -d'"' -f4)
    local type2=$(echo "$c2_get" | grep -o "\"what_id_type\":\"[^\"]*\"" | cut -d'"' -f4)
    
    local failures=0
    
    if [ "$type1" == "$p1_name" ]; then
        echo "  [OK] Event 1 correctly identified as $type1"
    else
        echo "  [FAIL] Event 1 type mismatch. Expected $p1_name, got $type1"
        failures=$((failures + 1))
    fi
    
    if [ "$type2" == "$p2_name" ]; then
        echo "  [OK] Event 2 correctly identified as $type2"
    else
        echo "  [FAIL] Event 2 type mismatch. Expected $p2_name, got $type2"
        failures=$((failures + 1))
    fi
    
    # 7. Cleanup
    echo "Cleaning up..."
    api_delete "/api/metadata/schemas/$child_name" > /dev/null
    api_delete "/api/metadata/schemas/$p1_name" > /dev/null
    api_delete "/api/metadata/schemas/$p2_name" > /dev/null
    
    if [ $failures -eq 0 ]; then
        test_passed "Polymorphic Lookups End-to-End"
    else
        test_failed "Polymorphic Verification Failed ($failures errors)" ""
    fi
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
