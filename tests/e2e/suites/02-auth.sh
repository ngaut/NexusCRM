#!/bin/bash
# tests/e2e/suites/02-auth.sh
# Authentication & Security Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Authentication & Security"

run_suite() {
    section_header "$SUITE_NAME"
    
    test_protected_endpoints
    test_login_validation
    test_successful_login
    test_current_user
}

test_protected_endpoints() {
    echo "Test 2.1: Protected Endpoints Require Auth"
    
    local response=$(api_get_unauth "/api/metadata/apps")
    assert_contains "$response" "Unauthorized" "Metadata endpoints properly protected"
    
    echo ""
    response=$(api_post_unauth "/api/formula/evaluate" '{"expression":"2+2","context":{}}')
    assert_contains "$response" "Unauthorized" "Formula endpoints properly protected"
    
    echo ""
    response=$(api_post_unauth "/api/data/query" '{"objectApiName":"Account"}')
    assert_contains "$response" "Unauthorized" "Data endpoints properly protected"
}

test_login_validation() {
    echo ""
    echo "Test 2.2: Login Validation"
    
    # Missing credentials
    local response=$(api_post_unauth "/api/auth/login" '{}')
    if echo "$response" | grep -qE "required|Bad Request"; then
        test_passed "Login validates required fields"
    else
        test_failed "Login validation (missing fields)" "$response"
    fi
    
    echo ""
    # Invalid email format
    response=$(api_post_unauth "/api/auth/login" '{"email":"notanemail","password":"test"}')
    if echo "$response" | grep -qE "Invalid email|Bad Request"; then
        test_passed "Login validates email format"
    else
        test_failed "Login validation (email format)" "$response"
    fi
    
    echo ""
    # Invalid credentials
    response=$(api_post_unauth "/api/auth/login" '{"email":"invalid@test.com","password":"wrongpassword"}')
    if echo "$response" | grep -qE "Unauthorized|Invalid"; then
        test_passed "Login rejects invalid credentials"
    else
        test_failed "Login validation (invalid creds)" "$response"
    fi
}

test_successful_login() {
    echo ""
    echo "Test 2.3: Successful Login"
    
    if api_login "$TEST_EMAIL" "$TEST_PASSWORD"; then
        test_passed "Login successful - token received"
        echo -e "  ${BLUE}Token:${NC} ${TOKEN:0:20}..."
        echo -e "  ${BLUE}User ID:${NC} $USER_ID"
    else
        test_failed "Login with valid credentials"
        echo -e "${RED}CRITICAL: Cannot proceed with authenticated tests${NC}"
        exit 1
    fi
}

test_current_user() {
    echo ""
    echo "Test 2.4: Get Current User Info (/me)"
    
    local response=$(api_get "/api/auth/me")
    if echo "$response" | grep -q '"id"' && echo "$response" | grep -q "$TEST_EMAIL"; then
        test_passed "GET /api/auth/me returns user info"
    else
        test_failed "GET /api/auth/me" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
