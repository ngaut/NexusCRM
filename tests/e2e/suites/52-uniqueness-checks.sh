#!/bin/bash
# tests/e2e/suites/52-uniqueness-checks.sh
# Uniqueness Constraint Validation Suite

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Uniqueness Constraint Validation"

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_object_label_uniqueness
    test_app_label_uniqueness
    test_group_label_uniqueness
    test_flow_name_uniqueness
}

test_object_label_uniqueness() {
    echo ""
    echo "Test 52.1: Object Label Uniqueness"
    local UNIQUE_ID=$(date +%s)
    local OBJ_LABEL="UniqueObj_$UNIQUE_ID"
    local OBJ_API_1="unique1_$UNIQUE_ID"
    local OBJ_API_2="unique2_$UNIQUE_ID"

    # Create First Object
    echo "  Creating first object: $OBJ_LABEL ($OBJ_API_1)"
    local res1=$(api_post "/api/metadata/objects" "{
        \"api_name\": \"$OBJ_API_1\",
        \"label\": \"$OBJ_LABEL\",
        \"plural_label\": \"${OBJ_LABEL}s\",
        \"description\": \"First object\",
        \"is_custom\": true,
        \"fields\": [
          {\"api_name\": \"name\", \"label\": \"Name\", \"type\": \"text\"}
        ]
    }")

    if echo "$res1" | grep -q "\"success\":true\|\"id\":"; then
        echo "  ✓ First object created successfully"
    else
         test_failed "Failed to create first object" "$res1"
         return 1
    fi

    # Create Second Object (Same Label, Different API Name)
    echo "  Creating second object with SAME label: $OBJ_LABEL ($OBJ_API_2)"
    # We expect this to FAIL
    local res2=$(curl -s -X POST "$BASE_URL/api/metadata/objects" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"api_name\": \"$OBJ_API_2\",
        \"label\": \"$OBJ_LABEL\",
        \"plural_label\": \"${OBJ_LABEL}s\",
        \"description\": \"Duplicate object\",
        \"is_custom\": true,
        \"fields\": [
          {\"api_name\": \"name\", \"label\": \"Name\", \"type\": \"text\"}
        ]
      }")

    # Expect 409 Conflict or Error message
    if echo "$res2" | grep -q "Conflict\|duplicate\|already exists\|Duplicate entry"; then
        test_passed "Duplicate Object Label Rejected"
    else
        test_failed "Duplicate Object Label was accepted!" "$res2"
    fi
}

test_app_label_uniqueness() {
    echo ""
    echo "Test 52.2: App Label Uniqueness"
    local UNIQUE_ID=$(date +%s)
    local APP_LABEL="UniqueApp_$UNIQUE_ID"
    local APP_ID_1="app1_$UNIQUE_ID"
    local APP_ID_2="app2_$UNIQUE_ID"

    echo "  Creating first app: $APP_LABEL"
    api_post "/api/metadata/apps" "{\"id\": \"$APP_ID_1\", \"label\": \"$APP_LABEL\", \"description\": \"First App\"}" > /dev/null
    echo "  ✓ First app created"

    echo "  Creating second app with SAME label..."
    # We use raw curl here to capture failure response directly if api_post masks it, 
    # but api_post returns the body so it's fine.
    local response_app=$(curl -s -X POST "$BASE_URL/api/metadata/apps" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"id\": \"$APP_ID_2\", \"label\": \"$APP_LABEL\", \"description\": \"Duplicate App\"}")

    if echo "$response_app" | grep -q "Conflict\|duplicate\|already exists\|Duplicate entry"; then
        test_passed "Duplicate App Label Rejected"
    else
        test_failed "Duplicate App Label was accepted!" "$response_app"
    fi
}

test_group_label_uniqueness() {
    echo ""
    echo "Test 52.3: Group Label Uniqueness"
    local UNIQUE_ID=$(date +%s)
    local GROUP_LABEL="UniqueGroup_$UNIQUE_ID"

    echo "  Creating first group..."
    api_post "/api/data/_System_Group" "{\"label\": \"$GROUP_LABEL\", \"name\": \"group1_$UNIQUE_ID\", \"type\": \"Regular\"}" > /dev/null
    echo "  ✓ First group created"

    echo "  Creating second group with SAME label..."
    local response_grp=$(curl -s -X POST "$BASE_URL/api/data/_System_Group" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"label\": \"$GROUP_LABEL\", \"name\": \"group2_$UNIQUE_ID\", \"type\": \"Regular\"}")

    if echo "$response_grp" | grep -q "Conflict\|duplicate\|already exists\|Duplicate entry"; then
        test_passed "Duplicate Group Label Rejected"
    else
        test_failed "Duplicate Group Label was accepted!" "$response_grp"
    fi
}

test_flow_name_uniqueness() {
    echo ""
    echo "Test 52.4: Flow Name Uniqueness"
    local UNIQUE_ID=$(date +%s)
    local FLOW_NAME="UniqueFlow_$UNIQUE_ID"

    echo "  Creating first flow..."
    api_post "/api/metadata/flows" "{\"name\": \"$FLOW_NAME\", \"label\": \"Flow 1\", \"trigger_type\": \"RecordChange\", \"trigger_object\": \"Account\", \"action_type\": \"create_record\", \"action_config\": {}, \"trigger_config\": {}}" > /dev/null
    echo "  ✓ First flow created"

    echo "  Creating second flow with SAME name..."
    local response_flow=$(curl -s -X POST "$BASE_URL/api/metadata/flows" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"name\": \"$FLOW_NAME\", \"label\": \"Flow 2\", \"trigger_type\": \"RecordChange\", \"trigger_object\": \"Lead\", \"action_type\": \"create_record\", \"action_config\": {}, \"trigger_config\": {}}")

    if echo "$response_flow" | grep -q "Conflict\|duplicate\|already exists\|Duplicate entry"; then
        test_passed "Duplicate Flow Name Rejected"
    else
        test_failed "Duplicate Flow Name was accepted!" "$response_flow"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
