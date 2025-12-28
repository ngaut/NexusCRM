#!/bin/bash
# tests/e2e/lib/helpers.sh
# Shared helper functions for E2E tests

# Get the directory of this file (use unique var to avoid clobbering caller's SCRIPT_DIR)
HELPERS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source config if not already loaded
if [ -z "$BASE_URL" ]; then
    source "$HELPERS_DIR/../config.sh"
fi

# Test result tracking
test_passed() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((TOTAL_PASSED++)) || true  # Ensure exit code is 0
}

test_failed() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    if [ -n "$2" ]; then
        echo -e "  ${YELLOW}Response:${NC} $2"
    fi
    ((TOTAL_FAILED++)) || true
}

section_header() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# JSON manipulation using jq (robust)
json_extract() {
    local response="$1"
    local field="$2"
    
    # Try direct field first, then try .record.field (API response wrapper)
    local result=$(echo "$response" | jq -r ".$field // .record.$field // empty" 2>/dev/null)
    echo "$result"
}

assert_json_has() {
    local response="$1"
    local field="$2"
    echo "$response" | grep -q "\"$field\""
}

assert_json_equals() {
    local response="$1"
    local field="$2"
    local expected="$3"
    echo "$response" | grep -q "\"$field\":\"$expected\""
}

# HTTP status code helpers
get_status_code() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"

    if [ "$method" = "GET" ]; then
        curl -s --max-time ${TIMEOUT:-30} -o /dev/null -w '%{http_code}' "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN"
    elif [ "$method" = "POST" ]; then
        curl -s --max-time ${TIMEOUT:-30} -o /dev/null -w '%{http_code}' -X POST "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data"
    elif [ "$method" = "PATCH" ]; then
        curl -s --max-time ${TIMEOUT:-30} -o /dev/null -w '%{http_code}' -X PATCH "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data"
    elif [ "$method" = "DELETE" ]; then
        curl -s --max-time ${TIMEOUT:-30} -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN"
    fi
}

# String helpers
assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-String contains check}"
    
    if echo "$haystack" | grep -q "$needle"; then
        test_passed "$message"
        return 0
    else
        test_failed "$message" "Expected to contain: $needle"
        return 1
    fi
}

assert_not_empty() {
    local value="$1"
    local message="${2:-Value not empty}"
    
    if [ -n "$value" ]; then
        test_passed "$message"
        return 0
    else
        test_failed "$message" "Value is empty"
        return 1
    fi
}

assert_status() {
    local actual="$1"
    local expected="$2"
    local message="${3:-HTTP status check}"
    
    if [ "$actual" = "$expected" ]; then
        test_passed "$message (HTTP $actual)"
        return 0
    else
        test_failed "$message" "Expected HTTP $expected, got $actual"
        return 1
    fi
}
