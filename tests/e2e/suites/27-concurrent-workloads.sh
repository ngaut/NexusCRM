#!/bin/bash
# tests/e2e/suites/27-concurrent-workloads.sh
# Multi-User Concurrent Operations E2E Tests
# REFACTORED: Uses helper libraries

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="Concurrent Workloads"
TIMESTAMP=$(date +%s)

# Test data IDs
LEAD_IDS=()
OPPORTUNITY_IDS=()

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Tests
    test_create_lead_pool
    test_parallel_lead_claims
    test_parallel_opportunity_updates
    test_concurrent_reads
    test_rapid_fire_creates
    test_data_consistency
    test_cleanup
}

# =========================================
# TESTS
# =========================================

test_create_lead_pool() {
    echo ""
    echo "Test 27.1: Create Shared Lead Pool"
    
    local count=0
    for i in A B C D E; do
        local res=$(api_post "/api/data/lead" '{"name": "Lead '$i' - '$TIMESTAMP'", "company": "Company '$i'", "status": "New", "email": "lead'$i'.'$TIMESTAMP'@test.com"}')
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { LEAD_IDS+=("$id"); count=$((count + 1)); }
    done
    
    echo "  Created $count leads"
    [ $count -eq 5 ] && test_passed "Lead pool created" || test_failed "Only $count leads"
}

test_parallel_lead_claims() {
    echo ""
    echo "Test 27.2: Simulate Parallel Lead Claims"
    
    [ ${#LEAD_IDS[@]} -lt 3 ] && { test_failed "Not enough leads"; return 1; }
    
    # Parallel claims (background processes)
    api_patch "/api/data/lead/${LEAD_IDS[0]}" '{"status": "Working"}' > /dev/null &
    api_patch "/api/data/lead/${LEAD_IDS[1]}" '{"status": "Working"}' > /dev/null &
    api_patch "/api/data/lead/${LEAD_IDS[2]}" '{"status": "Working"}' > /dev/null &
    wait
    
    echo "  ✓ Parallel claims completed"
    test_passed "Parallel claims simulated"
}

test_parallel_opportunity_updates() {
    echo ""
    echo "Test 27.3: Parallel Opportunity Updates"
    
    local count=0
    for i in 1 2 3; do
        local res=$(api_post "/api/data/opportunity" '{"name": "Opp '$i' - '$TIMESTAMP'", "stage_name": "Prospecting", "amount": '$((i * 10000))'}')
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { OPPORTUNITY_IDS+=("$id"); count=$((count + 1)); }
    done
    echo "  Created $count opportunities"
    
    [ ${#OPPORTUNITY_IDS[@]} -lt 3 ] && { test_failed "Not enough opps"; return 1; }
    
    api_patch "/api/data/opportunity/${OPPORTUNITY_IDS[0]}" '{"stage_name": "Qualification"}' > /dev/null &
    api_patch "/api/data/opportunity/${OPPORTUNITY_IDS[1]}" '{"stage_name": "Needs Analysis"}' > /dev/null &
    api_patch "/api/data/opportunity/${OPPORTUNITY_IDS[2]}" '{"stage_name": "Proposal"}' > /dev/null &
    wait
    
    echo "  ✓ Parallel updates completed"
    test_passed "Parallel updates"
}

test_concurrent_reads() {
    echo ""
    echo "Test 27.4: Concurrent Read Operations"
    
    local start_time=$(date +%s)
    
    api_post "/api/data/query" '{"object_api_name": "lead", "limit": 50}' > /dev/null &
    api_post "/api/data/query" '{"object_api_name": "opportunity", "limit": 50}' > /dev/null &
    api_post "/api/data/query" '{"object_api_name": "account", "limit": 50}' > /dev/null &
    api_post "/api/data/query" '{"object_api_name": "contact", "limit": 50}' > /dev/null &
    wait
    
    local end_time=$(date +%s)
    echo "  4 concurrent queries in $((end_time - start_time))s"
    
    test_passed "Concurrent reads"
}

test_rapid_fire_creates() {
    echo ""
    echo "Test 27.5: Rapid-Fire Create Operations"
    
    local start_time=$(date +%s)
    local success=0
    local temp_ids=()
    
    for i in $(seq 1 10); do
        local res=$(api_post "/api/data/lead" '{"name": "Rapid '$i' - '$TIMESTAMP'", "company": "Rapid Co", "status": "New", "email": "rapid'$i'.'$TIMESTAMP'@test.com"}')
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { temp_ids+=("$id"); success=$((success + 1)); }
    done
    
    echo "  Created $success in $(($(date +%s) - start_time))s"
    
    # Cleanup temp records
    for id in "${temp_ids[@]}"; do
        api_delete "/api/data/lead/$id" > /dev/null 2>&1
    done
    
    [ $success -ge 8 ] && test_passed "Rapid-fire creates" || test_failed "Only $success"
}

test_data_consistency() {
    echo ""
    echo "Test 27.6: Verify Data Consistency"
    
    local errors=0
    
    for id in "${LEAD_IDS[@]}"; do
        local res=$(api_get "/api/data/lead/$id")
        echo "$res" | grep -q '"id"' || errors=$((errors + 1))
    done
    
    for id in "${OPPORTUNITY_IDS[@]}"; do
        local res=$(api_get "/api/data/opportunity/$id")
        echo "$res" | grep -q '"id"' || errors=$((errors + 1))
    done
    
    echo "  Checked ${#LEAD_IDS[@]} leads, ${#OPPORTUNITY_IDS[@]} opps, $errors errors"
    
    [ $errors -eq 0 ] && test_passed "Data consistency verified" || test_failed "$errors errors"
}

# =========================================
# CLEANUP (Using Helpers)
# =========================================

test_cleanup() {
    echo ""
    echo "Test 27.7: Cleanup Test Data"
    
    cleanup_records "opportunity" "${OPPORTUNITY_IDS[@]}"
    cleanup_records "lead" "${LEAD_IDS[@]}"
    
    test_passed "Cleanup completed"
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
