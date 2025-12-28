#!/bin/bash
# tests/e2e/suites/01-infrastructure.sh
# Infrastructure & Deployment Tests

set +e  # Don't exit on error
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"

SUITE_NAME="Infrastructure & Deployment"

run_suite() {
    section_header "$SUITE_NAME"
    
    test_health_check
    test_go_server_process
    test_port_listener
    test_no_legacy_nodejs
}

test_health_check() {
    echo "Test 1.1: Health Check"
    local response=$(curl -s --max-time ${TIMEOUT:-30} "$BASE_URL/health")
    if echo "$response" | grep -q '"status":"ok"' && echo "$response" | grep -q '"server":"golang"'; then
        test_passed "Health endpoint returns correct Go backend response"
    else
        test_failed "Health endpoint" "$response"
    fi
}

test_go_server_process() {
    echo ""
    echo "Test 1.2: Verify Go Server Process"
    if ps aux | grep -E "backend/bin/server|/bin/server" | grep -v grep > /dev/null; then
        test_passed "Go server process is running"
    else
        test_failed "Go server process" "Process not found"
    fi
}

test_port_listener() {
    echo ""
    echo "Test 1.3: Verify Port 3001 Listener"
    if lsof -i:3001 | grep -q "LISTEN"; then
        PROCESS=$(lsof -i:3001 | grep LISTEN | awk '{print $1}')
        if [ "$PROCESS" = "server" ]; then
            test_passed "Port 3001 is listening (Go binary)"
        else
            test_failed "Port listener" "Wrong process: $PROCESS (expected: server)"
        fi
    else
        test_failed "Port 3001" "No listener found"
    fi
}

test_no_legacy_nodejs() {
    echo ""
    echo "Test 1.4: Verify No Legacy Node.js Backend"
    # Filter out: grep process itself, tsserver (IDE), and things running from node_modules
    if ps aux | grep -E "tsx.*server|node.*server" | grep -v grep | grep -v "tsserver" | grep -v "node_modules" > /dev/null; then
        NODE_PROC=$(ps aux | grep -E "tsx.*server|node.*server" | grep -v grep | grep -v "tsserver" | grep -v "node_modules")
        test_failed "No Node.js backend" "Found Node.js process: $NODE_PROC"
    else
        test_passed "No legacy Node.js/TypeScript server running"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
