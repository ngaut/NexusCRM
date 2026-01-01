#!/bin/bash
# tests/e2e/suites/49-ai-conversation-persistence.sh
# AI Conversation Persistence E2E Tests
# Verifies the /api/agent/conversation endpoints for save, load, and clear

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="AI Conversation Persistence Tests"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate first
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_get_empty_conversation
    test_save_conversation
    test_get_saved_conversation
    test_clear_conversation
}

test_get_empty_conversation() {
    echo "Test 49.1: Get Empty Conversation"
    
    response=$(api_get "/api/agent/conversation")
    
    if echo "$response" | grep -q '"messages"'; then
        echo "  ✓ Response contains 'messages' key"
    else
        test_failed "Response missing 'messages' key" "$response"
        return 1
    fi
    
    # Check that messages is empty or null
    messages=$(json_extract "$response" "messages")
    if [ "$messages" = "[]" ] || [ "$messages" = "null" ] || [ -z "$messages" ]; then
        test_passed "Empty conversation returned for new user"
    else
        echo "  ℹ️  Found existing conversation, will be overwritten by next test"
    fi
}

test_save_conversation() {
    echo "Test 49.2: Save Conversation"
    
    # Construct test messages
    local payload='{
        "messages": [
            {"role": "user", "content": "Hello AI"},
            {"role": "assistant", "content": "Hello! How can I help you today?"},
            {"role": "user", "content": "What is 2+2?"},
            {"role": "assistant", "content": "2+2 equals 4."}
        ],
        "title": "Test Conversation"
    }'
    
    response=$(api_post "/api/agent/conversation" "$payload")
    
    if echo "$response" | grep -q '"status"'; then
        status=$(json_extract "$response" "status")
        if [ "$status" = "created" ] || [ "$status" = "updated" ]; then
            test_passed "Conversation saved successfully (status: $status)"
        else
            test_failed "Unexpected save status: $status" "$response"
        fi
    else
        test_failed "Response missing 'status' key" "$response"
    fi
}

test_get_saved_conversation() {
    echo "Test 49.3: Get Saved Conversation"
    
    response=$(api_get "/api/agent/conversation")
    
    if echo "$response" | grep -q '"messages"'; then
        echo "  ✓ Response contains 'messages'"
    else
        test_failed "Response missing 'messages'" "$response"
        return 1
    fi
    
    # Check if our saved messages are there
    if echo "$response" | grep -q 'What is 2+2'; then
        test_passed "Saved messages persisted and retrieved correctly"
    else
        test_failed "Saved messages not found in response" "$response"
    fi
}

test_clear_conversation() {
    echo "Test 49.4: Clear Conversation"
    
    response=$(api_delete "/api/agent/conversation")
    
    if echo "$response" | grep -q '"status"'; then
        status=$(json_extract "$response" "status")
        if [ "$status" = "cleared" ]; then
            echo "  ✓ Conversation cleared"
        else
            test_failed "Unexpected clear status: $status" "$response"
            return 1
        fi
    fi
    
    # Verify conversation is now empty
    response=$(api_get "/api/agent/conversation")
    if echo "$response" | grep -q '\[\]'; then
        test_passed "Conversation cleared - messages are empty"
    else
        # Could show [] or empty messages - check for empty messages array
        messages=$(json_extract "$response" "messages")
        if [ "$messages" = "[]" ] || echo "$response" | grep -q '"messages":\[\]'; then
            test_passed "Conversation cleared - messages are empty"
        else
            test_failed "Conversation not properly cleared" "$response"
        fi
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
