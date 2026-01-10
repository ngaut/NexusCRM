#!/bin/bash
# tests/e2e/lib/api.sh
# API request wrapper functions

# Get the directory of this file (use unique var to avoid clobbering caller's SCRIPT_DIR)
API_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source config if not already loaded
if [ -z "$BASE_URL" ]; then
    source "$API_DIR/../config.sh"
fi

# Make authenticated GET request
api_get() {
    local endpoint="$1"
    curl -s --max-time ${TIMEOUT:-30} "$BASE_URL$endpoint" -H "Authorization: Bearer $TOKEN"
}

# Make authenticated POST request with JSON data
api_post() {
    local endpoint="$1"
    local data="$2"
    curl -s --max-time ${TIMEOUT:-30} -X POST "$BASE_URL$endpoint" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$data"
}

# Make authenticated PATCH request
api_patch() {
    local endpoint="$1"
    local data="$2"
    curl -s --max-time ${TIMEOUT:-30} -X PATCH "$BASE_URL$endpoint" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$data"
}

# Make authenticated DELETE request
api_delete() {
    local endpoint="$1"
    curl -s --max-time ${TIMEOUT:-30} -X DELETE "$BASE_URL$endpoint" \
        -H "Authorization: Bearer $TOKEN"
}

# Login and set TOKEN and USER_ID
api_login() {
    local email="${1:-$TEST_EMAIL}"
    local password="${2:-$TEST_PASSWORD}"
    
    local response=$(curl -v --max-time ${TIMEOUT:-30} -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"$password\"}")
    
    if echo "$response" | grep -q '"success":true' && echo "$response" | grep -q '"token"'; then
        export TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | sed 's/"token":"//')
        export USER_ID=$(echo "$response" | grep -o '"__sys_gen_id":"[^"]*' | head -1 | sed 's/"__sys_gen_id":"//')
        return 0
    else
        echo "Login failed: $response" >&2
        return 1
    fi
}

# Logout
api_logout() {
    curl -s --max-time ${TIMEOUT:-30} -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/auth/logout" \
        -H "Authorization: Bearer $TOKEN"
    export TOKEN=""
    export USER_ID=""
}

# Make unauthenticated request (for testing auth)
api_get_unauth() {
    local endpoint="$1"
    curl -v --max-time ${TIMEOUT:-30} "$BASE_URL$endpoint"
}

api_post_unauth() {
    local endpoint="$1"
    local data="$2"
    curl -s --max-time ${TIMEOUT:-30} -X POST "$BASE_URL$endpoint" \
        -H "Content-Type: application/json" \
        -d "$data"
}
