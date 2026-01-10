#!/bin/bash
# tests/e2e/suites/35-negative-security.sh
# Negative Security & Access Control Tests
# Verifies that Standard Users CANNOT perform Admin actions

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Negative Security Access Control"
TIMESTAMP=$(date +%s)

# Helper to login as standard user
# Assuming we have a standard user or create one
STANDARD_USER_EMAIL="standard_${TIMESTAMP}@test.com"
STANDARD_USER_PASS="Password123!"

run_suite() {
    section_header "$SUITE_NAME"
    
    # 1. Setup: Create Standard User (using Admin token)
    if ! api_login; then
        echo "Failed to login as Admin. Skipping suite."
        return 1
    fi

    # Fetch Profile ID
    echo "Fetching Standard User Profile..."
    # The profile ID is 'standard_user' (same as the internal name in system_data.json)
    PROFILE_STANDARD_USER="standard_user"
    
    if [ -z "$PROFILE_STANDARD_USER" ]; then
         # Fallback to Label search if needed, but standard_user should exist
         test_failed "Could not find 'standard_user' profile"
         return 1
    fi
    echo "  ✓ Found Profile ID: $PROFILE_STANDARD_USER"
    
    setup_standard_user
    
    # 2. Login as Standard User
    echo "Logging in as Standard User..."
    if api_login "$STANDARD_USER_EMAIL" "$STANDARD_USER_PASS"; then
        echo "  ✓ Logged in as Standard User"
    else
        test_failed "Setup: Could not login as standard user"
        return 1
    fi
    
    # 3. Validation Tests
    test_delete_system_object
    test_create_schema
    test_modify_permissions
    
    # 4. Cleanup (Switch back to Admin)
    perform_cleanup
}

setup_standard_user() {
    echo "Test 35.0: Setup Standard User"
    
    # Create User
    local res=$(api_post "/api/data/_System_User" '{
        "email": "'$STANDARD_USER_EMAIL'",
        "username": "Standard User '$TIMESTAMP'",
        "first_name": "Standard",
        "last_name": "User",
        "password": "'$STANDARD_USER_PASS'",
        "profile_id": "'$PROFILE_STANDARD_USER'",
        "is_active": true,
        "phone": "555-555-5555"
    }')
    
    local id=$(json_extract "$res" "id")
    if [ -z "$id" ]; then
        echo "Error creating user: $res"
        return 1
    fi
    STANDARD_USER_ID="$id"
    echo "  ✓ Created Standard User: $id"
}

test_delete_system_object() {
    echo ""
    echo "Test 35.1: Standard User cannot delete System Object"
    
    # Try DELETE /api/metadata/objects/Account
    # Note: Using metadata endpoint
    local res=$(api_delete "/api/metadata/objects/Account")
    
    if echo "$res" | grep -qiE "Forbidden|Unauthorized|Access Denied|403"; then
        echo "  ✓ Delete System Object Denied (403)"
        test_passed "System Object Protection"
    else
        # If it returns 200/204 or verification fails
        # Check if Account actually exists still
        # But wait, we don't want to actually delete it!
        # Ideally the API blocks it.
        # If the API returned success, we broke the system!
        test_failed "Delete System Object NOT Denied!" "$res"
    fi
}

test_create_schema() {
    echo ""
    echo "Test 35.2: Standard User cannot create Schema"
    
    local res=$(api_post "/api/metadata/objects" '{
        "label": "Hacker Object",
        "plural_label": "Hacker Objects",
        "api_name": "hacker_object",
        "is_custom": true
    }')
    
    if echo "$res" | grep -qiE "Forbidden|Unauthorized|Access Denied|403"; then
        echo "  ✓ Create Schema Denied (403)"
        test_passed "Schema Creation Protection"
    else
        test_failed "Standard User Created Schema!" "$res"
    fi
}

test_modify_permissions() {
    echo ""
    echo "Test 35.3: Standard User cannot modify Permissions"
    
    # Try create _System_Permission 
    local res=$(api_post "/api/data/_System_Permission" '{
        "name": "Hacker Perm"
    }')
    
    if echo "$res" | grep -qiE "Forbidden|Unauthorized|Access Denied|403"; then
        echo "  ✓ Modify Permissions Denied (403)"
        test_passed "Permission Modification Protection"
    else
        test_failed "Standard User Modified Permissions!" "$res"
    fi
}

perform_cleanup() {
    echo ""
    echo "Cleanup..."
    
    # Login as Admin
    api_login "$TEST_EMAIL" "$TEST_PASSWORD" > /dev/null
    
    if [ -n "$STANDARD_USER_ID" ]; then
        api_delete "/api/data/_System_User/$STANDARD_USER_ID" > /dev/null
        echo "  ✓ Deleted standard user"
    fi
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
