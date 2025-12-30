#!/bin/bash
# tests/e2e/suites/41-compact-feature.sh
# Compact Feature E2E Tests
# Verifies the /api/agent/compact endpoint

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Compact Feature Tests"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate first
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_compact_endpoint
}

test_compact_endpoint() {
    echo "Test 41.1: Compact Context Endpoint"
    
    # Construct a conversation payload with enough content
    # Construct a conversation payload with enough content to trigger compaction savings
    # Generate a long conversation to ensure summary is smaller than raw history
    local messages='[{"role": "system", "content": "You are a helpful assistant."}'
    for i in {1..20}; do
        messages+=",{\"role\": \"user\", \"content\": \"Tell me about Go features and why it is fast iteration $i\"}"
        messages+=",{\"role\": \"assistant\", \"content\": \"Go is a compiled, concurrently garbage-collected language that is known for its simplicity and efficiency. It has goroutines for concurrency. iteration $i\"}"
    done
    messages+=']'
    
    conversation_json="{\"messages\": $messages}"

    echo "  Sending POST /api/agent/compact..."
    response=$(api_post "/api/agent/compact" "$conversation_json")
    
    # 1. Check HTTP Status (implicitly handled by api_post failing mostly, but let's check content)
    # api_post implementation usually returns body. We check for keys.
    
    if echo "$response" | grep -q '"messages"'; then
         echo "  ✓ Response contains 'messages'"
    else
         test_failed "API Response missing 'messages'" "$response"
         return 1
    fi

    # 2. Extract Token Counts
    tokens_before=$(json_extract "$response" "tokens_before")
    tokens_after=$(json_extract "$response" "tokens_after")
    
    echo "  Tokens Before: $tokens_before"
    echo "  Tokens After:  $tokens_after"

    # 3. Validation Logic
    if [ -n "$tokens_before" ] && [ -n "$tokens_after" ]; then
        if [ "$tokens_after" -le "$tokens_before" ]; then
            test_passed "Compaction logic valid (After: $tokens_after <= Before: $tokens_before)"
        else
            test_failed "Tokens After ($tokens_after) > Tokens Before ($tokens_before) - Unexpected expansion" "$response"
        fi
    else
        test_failed "Missing token counts in response" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
