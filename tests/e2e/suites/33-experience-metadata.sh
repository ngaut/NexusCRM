#!/bin/bash
# tests/e2e/suites/33-experience-metadata.sh
# Experience Metadata Tests (Themes & Apps)

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Experience Metadata"

run_suite() {
    section_header "$SUITE_NAME"

    # Ensure we are admin
    if ! api_login "admin@test.com" "Admin123!"; then
        test_failed "Login as admin"
        return
    fi
    test_passed "Login as admin"

    test_theme_lifecycle
    test_object_theme_metadata
}

test_theme_lifecycle() {
    echo ""
    echo "Test 33.1: Theme Lifecycle (Create, Activate, Verify)"

    # 1. Create Theme
    local theme_name="E2E_Test_Theme_$(date +%s)"
    local payload="{\"name\": \"$theme_name\", \"colors\": {\"brand\": \"#ff00ff\"}, \"density\": \"compact\"}"
    
    local response=$(api_post "/api/metadata/themes" "$payload")
    
    if assert_json_has "$response" "theme"; then
         local theme_id=$(json_extract "$response" "id") # Extract from 'theme' object inside response?
         # HandleCreateEnvelope returns "theme": {...}
         # json_extract tries .record first...
         # Let's see helper implementation.
         # helper: .theme // .record.theme ...? No, logic is .$field // .record.$field.
         # So json_extract "$response" "id" matches .id or .record.id
         # But response is { "theme": { "__sys_gen_id": ... } }
         # So we need to extract from .theme.
         theme_id=$(echo "$response" | jq -r ".theme.id")
         
         if [ -n "$theme_id" ] && [ "$theme_id" != "null" ]; then
             test_passed "Create Theme '$theme_name' (ID: $theme_id)"
         else
             test_failed "Create Theme (ID not found in response)" "$response"
             return
         fi

         # 2. Activate Theme
         local activate_resp=$(curl -s -X PUT "$BASE_URL/api/metadata/themes/$theme_id/activate" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{}')
         if echo "$activate_resp" | grep -q "success"; then
             test_passed "Activate Theme $theme_id"
         else
             test_failed "Activate Theme" "$activate_resp"
         fi

         # 3. Verify Active Theme
         local get_resp=$(api_get "/api/metadata/themes/active")
         local active_name=$(echo "$get_resp" | jq -r ".theme.name")
         
         if [ "$active_name" == "$theme_name" ]; then
             test_passed "Verify Active Theme is '$theme_name'"
         else
             test_failed "Verify Active Theme" "Expected '$theme_name', got '$active_name'. Response: $get_resp"
         fi

    else
        test_failed "Create Theme" "$response"
    fi
}

test_object_theme_metadata() {
    echo ""
    echo "Test 33.2: Object Theme Metadata (theme_color)"

    # Verify 'Account' schema has 'theme_color' field in response
    # The API returns the *Schema* details. 'theme_color' is a property of the ObjectMetadata, not a field in list_fields?
    # Yes, it's a top-level property of the object definition.
    
    local response=$(api_get "/api/metadata/objects/Account")
    # Response structure: { "schema": { ... "theme_color": ... } }
    
    # We just check if the key exists in JSON, even if null
    if echo "$response" | grep -q "theme_color"; then
         test_passed "Account schema contains 'theme_color' property"
    else
         test_failed "Account schema missing 'theme_color'" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
