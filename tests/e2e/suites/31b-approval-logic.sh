#!/bin/bash
# tests/e2e/suites/31b-approval-logic.sh
# Approval Process Logic Tests (Self-Contained)

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Approval Process Logic (Self-Contained)"

run_suite() {
    section_header "$SUITE_NAME"
    
    # Authenticate first
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_check_process_endpoint
    test_approval_submission_flow
}

# Test 31b.1: Verify CheckProcess Endpoint
test_check_process_endpoint() {
    echo "Test 31b.1: Verify CheckProcess Endpoint"
    
    # We assume 'Lead' has an active process call "Lead Test Approval" from previous manual setup
    # If not, we should probably create one, but let's first check if ANY process exists
    
    # GET /api/approvals/check/:objectApiName
    response=$(api_get "/api/approvals/check/Lead")
    
    # We verify the structure of the response
    # { "has_process": true/false, "process_name": "..." }
    
    if echo "$response" | grep -q '"has_process"'; then
        test_passed "GET /api/approvals/check/Lead returns valid structure"
    else
        test_failed "GET /api/approvals/check/Lead failed" "$response"
    fi
}

# Test 31b.2: End-to-End Submission Flow
test_approval_submission_flow() {
    echo "Test 31b.2: End-to-End Submission Flow"

    # Step 1: Create a Lead to submit
    echo "  1. Creating test Lead..."
    lead_resp=$(api_post "/api/data/Lead" '{"name": "Approval E2E Candidate", "company": "Test Corp", "status": "New", "email": "candidate@test.com"}')
    lead_id=$(json_extract "$lead_resp" "id")
    
    if [ -z "$lead_id" ] || [ "$lead_id" == "null" ]; then
        test_failed "Failed to create Lead for approval test" "$lead_resp"
        return 1
    fi
    test_passed "Created Lead: $lead_id"

    # Step 2: Ensure an Approval Process exists for Lead
    # We will create one strictly for this test to be self-contained
    echo "  2. Ensuring Approval Process exists..."
    # Check current processes
    proc_check=$(api_post "/api/data/query" '{"object_api_name": "_System_ApprovalProcess", "criteria": [{"field": "object_api_name", "op": "=", "val": "Lead"}, {"field": "is_active", "op": "=", "val": true}]}')
    
    if ! echo "$proc_check" | grep -q '"id"'; then
        echo "     No active process found. Creating one..."
        TIMESTAMP=$(date +%s)
        create_proc_resp=$(api_post "/api/data/_System_ApprovalProcess" '{
            "name": "E2E Auto Process '$TIMESTAMP'", 
            "object_api_name": "Lead", 
            "is_active": true, 
            "approver_type": "Self",
            "description": "Created by E2E Test"
        }')
        proc_id=$(json_extract "$create_proc_resp" "id")
        if [ -n "$proc_id" ]; then
            test_passed "Created temporary Approval Process: $proc_id"
        else
            test_failed "Failed to create Approval Process" "$create_proc_resp"
            return 1
        fi
    else
        test_passed "Active Approval Process already exists"
    fi

    # Step 3: Submit for Approval
    echo "  3. Submitting Lead ($lead_id) for approval..."
    # Endpoint: POST /api/approvals/submit
    # Payload: { "record_id": "...", "object_api_name": "Lead", "comment": "..." }
    submit_resp=$(api_post "/api/approvals/submit" "{\"record_id\": \"$lead_id\", \"object_api_name\": \"Lead\", \"comment\": \"E2E Test Submission\"}")
    
    # Check for success (200 OK and work_item_id)
    if echo "$submit_resp" | grep -q '"work_item_id"'; then
        work_item_id=$(json_extract "$submit_resp" "work_item_id")
        test_passed "Submission successful. Work Item ID: $work_item_id"
        
        # Verify Work Item exists in DB
        wi_check=$(api_get "/api/data/_System_ApprovalWorkItem/$work_item_id")
        if echo "$wi_check" | grep -q "$lead_id"; then
             test_passed "Verified Work Item exists in database and links to Lead"
        else
             test_failed "Work Item not found in DB or malformed" "$wi_check"
        fi

        # Verify Work Item visibility in Pending Queue (Critical Regression Test for Self Approval)
        pending_resp=$(api_get "/api/approvals/pending")
        if echo "$pending_resp" | grep -q "$work_item_id"; then
             test_passed "Verified Work Item is visible in Pending Queue (Self Approval logic works)"
        else
             test_failed "Work Item NOT found in Pending Queue - Potential Regression!" "$pending_resp"
        fi
        
    else
        test_failed "Submission failed (Potential 500 Error Regression)" "$submit_resp"
        return 1
    fi

    # Cleanup
    echo "  4. Cleanup..."
    api_delete "/api/data/Lead/$lead_id" > /dev/null
    # We assume we shouldn't delete the process as it might be used by other tests or users, 
    # but strictly speaking a test should clean up. For now we leave it if it was pre-existing.
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
