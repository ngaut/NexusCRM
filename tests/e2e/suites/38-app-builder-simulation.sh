#!/bin/bash
# tests/e2e/suites/38-app-builder-simulation.sh
# App Builder Simulation
# Simulates creating a Custom Object, adding fields, and using it.

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="App Builder Simulation"
TIMESTAMP=$(date +%s)
OBJ_NAME="Project_$TIMESTAMP"
API_NAME="project_$TIMESTAMP"

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_create_custom_object
    test_add_custom_fields
    test_create_record
    test_schema_evolution
    test_cleanup
}

test_create_custom_object() {
    echo ""
    echo "Test 38.1: Create Custom Object ($OBJ_NAME)"
    
    local res=$(api_post "/api/metadata/objects" '{
        "label": "'$OBJ_NAME'",
        "plural_label": "'$OBJ_NAME's",
        "api_name": "'$API_NAME'",
        "is_custom": true,
        "description": "A simulated custom object for Projects"
    }')
    
    if assert_contains "$res" "$API_NAME" "Object Created"; then
        echo "  ✓ Custom Object '$OBJ_NAME' created successfully"
    else
        test_failed "Failed to create custom object" "$res"
        return 1
    fi
}

test_add_custom_fields() {
    echo ""
    echo "Test 38.2: Add Fields (Date, Currency, Lookup)"
    
    # 1. Start Date (Date)
    local res1=$(api_post "/api/metadata/objects/$API_NAME/fields" '{
        "label": "Start Date",
        "api_name": "start_date",
        "type": "Date"
    }')
    assert_contains "$res1" "start_date" "Date Field Created"
    
    # 2. Budget (Currency)
    local res2=$(api_post "/api/metadata/objects/$API_NAME/fields" '{
        "label": "Budget",
        "api_name": "budget",
        "type": "Currency"
    }')
    assert_contains "$res2" "budget" "Currency Field Created"
    
    # 3. Manager (Lookup to User)
    local res3=$(api_post "/api/metadata/objects/$API_NAME/fields" '{
        "label": "Manager",
        "api_name": "manager",
        "type": "Lookup",
        "reference_to": ["_System_User"]
    }')
    assert_contains "$res3" "manager" "Lookup Field Created"
}

test_create_record() {
    echo ""
    echo "Test 38.3: Create Record of Custom Object"
    
    # Get current user ID for lookup
    local me=$(api_get "/api/auth/me")
    local user_id=$(json_extract "$me" "id")
    if [ -z "$user_id" ]; then user_id=$(json_extract "$me" "user_id"); fi
    
    local payload='{
        "name": "Alpha Project",
        "start_date": "2025-01-01T00:00:00Z",
        "budget": 50000,
        "manager": "'$user_id'"
    }'
    
    local res=$(api_post "/api/data/$API_NAME" "$payload")
    local id=$(json_extract "$res" "id")
    
    if [ -n "$id" ]; then
        echo "  ✓ Created Record ID: $id"
        test_passed "Record Creation"
        
        # Verify Data
        local check=$(api_get "/api/data/$API_NAME/$id")
        assert_contains "$check" "Alpha Project" "Name persisted"
        assert_contains "$check" "50000" "Currency persisted"
        
        TEST_RECORD_ID="$id"
    else
        test_failed "Failed to create record" "$res"
    fi
}

test_schema_evolution() {
    echo ""
    echo "Test 38.4: Schema Evolution (Add Field & Update Record)"
    
    # Add 'Status' Picklist
    local res=$(api_post "/api/metadata/objects/$API_NAME/fields" '{
        "label": "Project Status",
        "api_name": "status",
        "type": "Picklist",
        "options": ["Planned", "Active", "Completed"]
    }')
    
    if assert_contains "$res" "status" "Picklist Field Added"; then
        echo "  ✓ Schema Evolved"
        
        # Update Record with new field
        if [ -n "$TEST_RECORD_ID" ]; then
            local update=$(api_patch "/api/data/$API_NAME/$TEST_RECORD_ID" '{"status": "Active"}')
            local check=$(api_get "/api/data/$API_NAME/$TEST_RECORD_ID")
            assert_contains "$check" "Active" "New Field Data Persisted"
        fi
    fi
}

test_cleanup() {
    # Delete the custom object definition (Cascade should handle data, though strictly we might need to delete data first if configured so)
    # Our system supports cascading schema deletion? Let's try.
    # Usually safer to just delete schema.
    
    # The API might block schema delete if data exists? 
    # Current implementation: DeleteSchema checks permissions but logic might vary. 
    # Let's assume it allows it or we leave it (it's a test environment w/ wipe_db usually).
    
    echo "Cleanup skipped (relying on wipe_db or subsequent runs)"
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
