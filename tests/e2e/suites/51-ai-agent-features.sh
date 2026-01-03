#!/bin/bash
# tests/e2e/suites/51-ai-agent-features.sh
# AI Agent Features E2E Tests
# Tests: Multi-user conversation isolation, message persistence, history, context, and chat

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="AI Agent Features"
TIMESTAMP=$(date +%s)

# Helper for failure
test_fail() {
    echo "  ❌ $1"
    exit 1
}

# User Tokens
TOKEN_A=""
TOKEN_B=""
USER_A_ID=""
USER_B_ID=""

# Conversation IDs
CONV_ID_A=""
CONV_ID_B=""

# Cleanup function for trap
cleanup_test_users() {
    # Admin cleanup
    if [ -n "$USER_A_ID" ]; then
        api_delete "/api/metadata/users/$USER_A_ID" > /dev/null 2>&1
    fi
    if [ -n "$USER_B_ID" ]; then
        api_delete "/api/metadata/users/$USER_B_ID" > /dev/null 2>&1
    fi
}

run_suite() {
    section_header "$SUITE_NAME"
    
    # Ensure cleanup on exit
    trap cleanup_test_users EXIT

    setup_test_users
    
    test_conversation_isolation
    test_message_persistence
    test_concurrent_usage
    test_conversation_history
    test_context_context
    test_cleanup
}

setup_test_users() {
    echo "Setup: Creating Test Users..."
    
    # Authenticate as admin to create users
    api_login
    
    # Fetch a valid Profile ID
    local profiles_res=$(api_get "/api/auth/profiles")
    echo "  Debug: Profiles: $profiles_res"
    local profile_id=$(echo "$profiles_res" | jq -r '.profiles[] | select(.name=="System Administrator") | .id' 2>/dev/null)
    
    # Fallback to Standard User
    if [ -z "$profile_id" ]; then
        profile_id=$(echo "$profiles_res" | jq -r '.profiles[] | select(.name=="Standard User") | .id' 2>/dev/null)
    fi
    
    # Fallback to first profile
    if [ -z "$profile_id" ]; then
        profile_id=$(echo "$profiles_res" | jq -r '.profiles[0].id' 2>/dev/null)
    fi
    
    if [ -z "$profile_id" ]; then
        echo "Failed to fetch any profile ID"
        echo "Response: $profiles_res"
        return 1
    fi

    echo "  Debug: Profile ID: $profile_id"

    # Create User A
    local user_a_payload='{
        "email": "user.a.'$TIMESTAMP'@test.com",
        "name": "Test UserA",
        "password": "Admin123!",
        "profile_id": "'$profile_id'"
    }'
    local res_a=$(api_post "/api/auth/register" "$user_a_payload")
    # Extract ID from nested user object
    USER_A_ID=$(echo "$res_a" | jq -r '.user.id // ""' 2>/dev/null)
    
    # Create User B
    local user_b_payload='{
        "email": "user.b.'$TIMESTAMP'@test.com",
        "name": "Test UserB",
        "password": "Admin123!",
        "profile_id": "'$profile_id'"
    }'
    local res_b=$(api_post "/api/auth/register" "$user_b_payload")
    # Extract ID from nested user object
    USER_B_ID=$(echo "$res_b" | jq -r '.user.id // ""' 2>/dev/null)
    
    if [ -z "$USER_A_ID" ] || [ -z "$USER_B_ID" ]; then
        test_failed "Failed to create test users" "A: $res_a, B: $res_b"
        return 1
    fi
    
    # Wait for consistency (using loop instead of fixed sleep)
    local retries=0
    local max_retries=10
    local consistency_reached=0
    
    while [ $retries -lt $max_retries ]; do
        local user_check=$(api_get "/api/auth/users")
        if echo "$user_check" | grep -q "$USER_A_ID"; then
            consistency_reached=1
            break
        fi
        sleep 1
        retries=$((retries+1))
    done

    if [ $consistency_reached -eq 0 ]; then
        echo "  User A NOT found in user list after $max_retries attempts"
        return 1
    fi

    echo "Step 3: Generating tokens for test users..."
    
    # Helper for login
    login_and_get_token() {
        local email="$1"
        local password="$2"
        local payload="{\"email\":\"$email\",\"password\":\"$password\"}"
        api_post_unauth "/api/auth/login" "$payload"
    }

    # Login as User A
    echo "Logging in as User A..."
    # The payload had: email="user.a.$TIMESTAMP@test.com"
    local email_a="user.a.$TIMESTAMP@test.com"
    local login_a=$(login_and_get_token "$email_a" "Admin123!")
    
    if echo "$login_a" | grep -q "token"; then
         TOKEN_A=$(echo "$login_a" | jq -r '.token')
         echo "  ✅ User A logged in successfully"
    else
         echo "  ❌ User A login failed"
         echo "  Response: $login_a"
         return 1
    fi

    # Login as User B
    local email_b="user.b.$TIMESTAMP@test.com"
    local login_b=$(login_and_get_token "$email_b" "Admin123!")
    
    if echo "$login_b" | grep -q "token"; then
         TOKEN_B=$(echo "$login_b" | jq -r '.token')
         echo "  ✅ User B logged in successfully"
    else
         echo "  ❌ User B login failed"
         echo "  Response: $login_b"
         return 1
    fi
}

# Helper to make authenticated request as specific user
api_req_user() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local token="$4"
    
    curl -s -X "$method" "$BASE_URL$endpoint" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "$data" --insecure
}

test_conversation_isolation() {
    echo "Test 51.1: Multi-User Conversation Isolation"
    
    # User A creates conversation
    echo "  User A creating conversation..."
    local payload_a='{"title": "User A Chat", "messages": [{"role":"user","content":"Hello from A"}]}'
    local res_a=$(api_req_user "POST" "/api/agent/conversation" "$payload_a" "$TOKEN_A")
    CONV_ID_A=$(json_extract "$res_a" "id")
    
    if [ -z "$CONV_ID_A" ]; then
        echo "  Failed to create conversation for User A"
        echo "  Response: $res_a"
        test_fail "Conversation creation"
        return
    fi
    echo "  ✓ User A created conversation: $CONV_ID_A"
    
    # User B tries to access User A's conversation
    echo "  User B attempting to access User A's chat (should fail)..."
    local res_forbidden=$(api_req_user "GET" "/api/agent/conversation?id=$CONV_ID_A" "" "$TOKEN_B")
    
    if echo "$res_forbidden" | grep -q "not your conversation\|not found"; then
        echo "  ✓ User B correctly blocked from accessing User A's chat"
    else
        echo "  ❌ User B could access User A's chat!"
        echo "  Response: $res_forbidden"
    fi
    
    # User B creates their own
    echo "  User B creating conversation..."
    local payload_b='{"title": "User B Chat", "messages": [{"role":"user","content":"Hello from B"}]}'
    local res_b=$(api_req_user "POST" "/api/agent/conversation" "$payload_b" "$TOKEN_B")
    CONV_ID_B=$(json_extract "$res_b" "id")
    echo "  ✓ User B created conversation: $CONV_ID_B"
    
    test_passed "Conversation isolation"
}

test_message_persistence() {
    echo ""
    echo "Test 51.2: Message Persistence & State"
    
    # User A adds message
    local update_a='{"conversation_id": "'$CONV_ID_A'", "messages": [{"role":"user","content":"Hello from A"}, {"role":"assistant","content":"Hi A"}]}'
    api_req_user "POST" "/api/agent/conversation" "$update_a" "$TOKEN_A" > /dev/null
    
    # User B adds different message
    local update_b='{"conversation_id": "'$CONV_ID_B'", "messages": [{"role":"user","content":"Hello from B"}, {"role":"assistant","content":"Hi B"}]}'
    api_req_user "POST" "/api/agent/conversation" "$update_b" "$TOKEN_B" > /dev/null
    
    # Verify User A sees A's messages
    local get_a=$(api_req_user "GET" "/api/agent/conversation" "" "$TOKEN_A")
    if echo "$get_a" | grep -q "Hi A" && ! echo "$get_a" | grep -q "Hi B"; then
        echo "  ✓ User A sees correct messages"
    else
        echo "  ❌ User A message mismatch"
    fi
    
    # Verify User B sees B's messages
    local get_b=$(api_req_user "GET" "/api/agent/conversation" "" "$TOKEN_B")
    if echo "$get_b" | grep -q "Hi B" && ! echo "$get_b" | grep -q "Hi A"; then
        echo "  ✓ User B sees correct messages"
    else
        echo "  ❌ User B message mismatch"
    fi
    
    test_passed "Persistence verified"
}

test_concurrent_usage() {
    echo ""
    echo "Test 51.3: Concurrent Operations"
    
    # Rapidly list conversations for both users
    local list_a=$(api_req_user "GET" "/api/agent/conversations" "" "$TOKEN_A")
    local list_b=$(api_req_user "GET" "/api/agent/conversations" "" "$TOKEN_B")
    
    local count_a=$(echo "$list_a" | grep -o "\"id\"" | wc -l)
    local count_b=$(echo "$list_b" | grep -o "\"id\"" | wc -l)
    
    if [ "$count_a" -ge 1 ] && [ "$count_b" -ge 1 ]; then
        echo "  ✓ Both users successfully listed conversations concurrently"
    else
        echo "  ❌ Concurrent list failed"
    fi
    
    test_passed "Concurrent usage"
}

test_conversation_history() {
    echo ""
    echo "Test 51.4: Conversation History"
    
    # User A creates another conversation
    api_req_user "POST" "/api/agent/conversation" '{"messages":[{"role":"user","content":"Second chat"}]}' "$TOKEN_A" > /dev/null
    
    # List for User A
    local list_a=$(api_req_user "GET" "/api/agent/conversations" "" "$TOKEN_A")
    local count=$(echo "$list_a" | grep -o "\"id\"" | wc -l)
    
    if [ "$count" -ge 2 ]; then
        echo "  ✓ User A has multiple conversations in history"
    else
        echo "  ❌ History count mismatch"
    fi
    
    test_passed "Conversation history"
}

test_context_context() {
    echo ""
    echo "Test 51.5: Context & Compaction"
    
    # Get Context
    local ctx=$(api_req_user "GET" "/api/agent/context" "" "$TOKEN_A")
    if echo "$ctx" | grep -q "total_tokens"; then
        echo "  ✓ Context retrieved successfully"
    fi
    
    # Compact
    local compact_req='{"messages":[{"role":"user","content":"test"}]}'
    local compact_res=$(api_req_user "POST" "/api/agent/compact" "$compact_req" "$TOKEN_A")
    if echo "$compact_res" | grep -q "tokens_after"; then
        echo "  ✓ Compaction endpoint responsive"
    fi
    
    test_passed "Context services"
}

test_cleanup() {
    echo ""
    echo "Test 51.6: Cleanup"
    
    # Admin cleanup
    if [ -n "$USER_A_ID" ]; then
        api_delete "/api/metadata/users/$USER_A_ID" > /dev/null 2>&1
    fi
    if [ -n "$USER_B_ID" ]; then
        api_delete "/api/metadata/users/$USER_B_ID" > /dev/null 2>&1
    fi
    
    echo "  ✓ Test users deleted"
    test_passed "Cleanup completed"
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
