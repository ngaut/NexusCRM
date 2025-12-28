#!/bin/bash
# tests/e2e/suites/36-auth-edge-cases.sh
# Authentication Edge Cases
# Tests Inactive Users, Token Expiry simulation

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Auth Edge Cases"
TIMESTAMP=$(date +%s)

INACTIVE_USER_EMAIL="inactive_${TIMESTAMP}@test.com"
INACTIVE_USER_PASS="Password123!"

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    # Fetch Profile ID
    local profile_res=$(api_post "/api/data/query" '{"object_api_name": "_System_Profile", "filters": [{"field": "name", "operator": "=", "value": "standard_user"}]}')
    PROFILE_STANDARD_USER=$(json_extract "$profile_res" "records[0].id")
    
    if [ -z "$PROFILE_STANDARD_USER" ]; then
         test_failed "Could not find 'standard_user' profile"
         return 1
    fi

    test_inactive_user_login
    test_invalid_token_usage
}

test_inactive_user_login() {
    echo ""
    echo "Test 36.1: Inactive User Login Blocked"
    
    # 1. Create User with is_active=false
    echo "  Creating Inactive User..."
    local res=$(api_post "/api/data/_System_User" '{
        "email": "'$INACTIVE_USER_EMAIL'",
        "username": "Inactive User '$TIMESTAMP'",
        "first_name": "Inactive",
        "last_name": "User",
        "password": "'$INACTIVE_USER_PASS'",
        "profile_id": "'$PROFILE_STANDARD_USER'",
        "is_active": false,
        "phone": "555-555-5555"
    }')
    
    local id=$(json_extract "$res" "id")
    if [ -z "$id" ]; then
        test_failed "Setup: Create Inactive User Failed" "$res"
        return
    fi
    
    # 2. Attempt Login
    echo "  Attempting Login..."
    local login_res=$(api_post_unauth "/api/auth/login" '{
        "email": "'$INACTIVE_USER_EMAIL'",
        "password": "'$INACTIVE_USER_PASS'"
    }')
    
    if echo "$login_res" | grep -qiE "inactive|disabled|locked|Unauthorized|401"; then
        echo "  ✓ Inactive User Login Denied"
        test_passed "Inactive User Protection"
    else
        test_failed "Inactive User Successfully Logged In!" "$login_res"
    fi
    
    # Cleanup
    api_login > /dev/null # Re-login as Admin
    api_delete "/api/data/_System_User/$id" > /dev/null
}

test_invalid_token_usage() {
    echo ""
    echo "Test 36.2: Invalid/Expired Token Usage"
    
    # Use garbage token
    local saved_token="$TOKEN"
    TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.garbage.garbage"
    
    local res=$(api_get "/api/auth/me")
    
    if echo "$res" | grep -qiE "Unauthorized|Invalid token|401"; then
        echo "  ✓ Invalid Token Rejected"
        test_passed "Token Validation Enforced"
    else
        test_failed "Invalid Token Accepted!" "$res"
    fi
    
    TOKEN="$saved_token"
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
