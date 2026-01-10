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
    
    # API consistently returns data in .data wrapper
    local result=$(echo "$response" | jq -r ".data.$field // .$field // empty" 2>/dev/null)
    
    # If field is "id" and empty, try __sys_gen_id instead (UUID migration compatibility)
    if [ -z "$result" ] && [ "$field" = "id" ]; then
        result=$(echo "$response" | jq -r ".data.__sys_gen_id // .__sys_gen_id // empty" 2>/dev/null)
    fi
    
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

# Cleanup helper: Delete items where a field starts with a prefix
delete_items_by_prefix() {
    local endpoint="$1"
    local field="$2"
    local prefix="$3"
    local description="$4"

    echo "  Cleaning up $description ($field starts with '$prefix')..."
    
    # fetch all items
    local response=$(api_get "$endpoint")
    
    # Extract IDs of items matching prefix
    # Use jq to select items where the field starts with prefix, then output ID
    local ids=$(echo "$response" | jq -r ".data[] | select(.$field | tostring | startswith(\"$prefix\")) | .__sys_gen_id // .id")
    
    if [ -z "$ids" ]; then
        echo "    No items found to clean."
        return
    fi

    for id in $ids; do
        if [ "$id" != "null" ] && [ -n "$id" ]; then
            echo "    Deleting $id..."
            api_delete "$endpoint/$id" > /dev/null
        fi
    done
}

# Cleanup helper: Delete items VIA QUERY command (for system objects without list endpoints)
delete_via_query_by_prefix() {
    local object_api_name="$1"
    local field="$2"
    local prefix="$3"
    local description="$4"
    local delete_endpoint="${5:-/api/data/$object_api_name}"

    echo "  Cleaning up $description ($field starts with '$prefix') via Query..."
    
    # Query for items without 'fields' param (returns all fields, safe)
    # limit=1000 to ensure we catch recent items even if pagination is active
    local response=$(api_post "/api/data/query" "{\"object_api_name\": \"$object_api_name\", \"limit\": 1000}")

    # Extract IDs
    local ids=$(echo "$response" | jq -r ".data[] | select(.$field | tostring | startswith(\"$prefix\")) | .__sys_gen_id // .id")
    
    if [ -z "$ids" ]; then
        echo "    No items found to clean."
        return
    fi

    for id in $ids; do
        if [ "$id" != "null" ] && [ -n "$id" ]; then
            echo "    Deleting $id..."
            api_delete "$delete_endpoint/$id" > /dev/null
        fi
    done
}
