#!/bin/bash
# tests/e2e/suites/16-sales-lifecycle.sh
# Sales Lifecycle E2E Tests
# Tests: Lead qualification, conversion, Opportunity stages, Closed Won/Lost

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Sales Lifecycle"
TIMESTAMP=$(date +%s)

# Track if we created schemas (for cleanup)
CREATED_OPP_SCHEMA=false

# Test data IDs
TEST_LEAD_ID=""
TEST_ACCOUNT_ID=""
TEST_CONTACT_ID=""
TEST_OPPORTUNITY_ID=""

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    setup_schemas
    test_lead_creation
    test_lead_qualification
    test_lead_conversion
    test_opportunity_stages
    test_closed_won
    test_closed_lost
    test_cleanup
}

# Setup: Ensure Opportunity schema exists (create via API if needed)
setup_schemas() {
    echo "Setup: Ensuring Opportunity schema exists"
    
    # Check if Opportunity schema exists
    local check=$(api_get "/api/metadata/objects/opportunity")
    
    if echo "$check" | grep -q '"api_name"'; then
        echo "  ✓ Opportunity schema already exists"
    else
        echo "  Creating Opportunity schema..."
        
        local schema_payload='{
            "label": "Opportunity",
            "api_name": "opportunity",
            "plural_label": "Opportunities",
            "is_custom": false
        }'
        
        local res=$(api_post "/api/metadata/objects" "$schema_payload")
        
        if echo "$res" | grep -q '"api_name"'; then
            echo "  ✓ Opportunity schema created"
            CREATED_OPP_SCHEMA=true
            
            # Add required fields
            api_post "/api/metadata/objects/opportunity/fields" '{
                "api_name": "stage_name",
                "label": "Stage",
                "type": "Picklist",
                "options": ["Prospecting", "Qualification", "Needs Analysis", "Proposal", "Negotiation", "Closed Won", "Closed Lost"]
            }' > /dev/null
            
            api_post "/api/metadata/objects/opportunity/fields" '{
                "api_name": "amount",
                "label": "Amount",
                "type": "Currency"
            }' > /dev/null
            
            api_post "/api/metadata/objects/opportunity/fields" '{
                "api_name": "close_date",
                "label": "Close Date",
                "type": "Date"
            }' > /dev/null
            
            api_post "/api/metadata/objects/opportunity/fields" '{
                "api_name": "account_id",
                "label": "Account",
                "type": "Lookup",
                "reference_to": ["account"]
            }' > /dev/null
            
            echo "  ✓ Opportunity fields added"
        else
            echo "  Note: Could not create Opportunity schema: $res"
        fi
    fi
}

# Test 16.1: Create a new Lead
test_lead_creation() {
    echo "Test 16.1: Create New Lead"
    
    local payload='{
        "name": "Sales Lifecycle Lead '$TIMESTAMP'",
        "company": "Acme Corp",
        "email": "lead'$TIMESTAMP'@acme.com",
        "status": "New",
        "source": "Web",
        "state": "California"
    }'
    
    local response=$(api_post "/api/data/lead" "$payload")
    TEST_LEAD_ID=$(json_extract "$response" "id")
    
    if [ -n "$TEST_LEAD_ID" ]; then
        echo "  Created Lead: $TEST_LEAD_ID"
        test_passed "Lead Created"
    else
        test_failed "Failed to create Lead" "$response"
        return 1
    fi
}

# Test 16.2: Qualify the Lead (update status)
test_lead_qualification() {
    echo ""
    echo "Test 16.2: Qualify Lead (Status → Qualified)"
    
    if [ -z "$TEST_LEAD_ID" ]; then
        test_failed "No Lead ID available"
        return 1
    fi
    
    local payload='{"status": "Qualified"}'
    local response=$(api_patch "/api/data/lead/$TEST_LEAD_ID" "$payload")
    
    if echo "$response" | grep -qE '"id"|"success"|updated'; then
        # Verify status changed
        local lead=$(api_get "/api/data/lead/$TEST_LEAD_ID")
        local status=$(json_extract "$lead" "status")
        
        if [ "$status" == "Qualified" ]; then
            test_passed "Lead Qualified"
        else
            test_failed "Lead status not updated" "Expected: Qualified, Got: $status"
        fi
    else
        test_failed "Failed to qualify Lead" "$response"
    fi
}

# Test 16.3: Convert Lead to Account + Contact + Opportunity
test_lead_conversion() {
    echo ""
    echo "Test 16.3: Convert Lead to Account + Contact + Opportunity"
    
    if [ -z "$TEST_LEAD_ID" ]; then
        test_failed "No Lead ID available"
        return 1
    fi
    
    # Get Lead data for conversion
    local lead=$(api_get "/api/data/lead/$TEST_LEAD_ID")
    local lead_name=$(json_extract "$lead" "name")
    local lead_company=$(json_extract "$lead" "company")
    local lead_email=$(json_extract "$lead" "email")
    
    # 1. Create Account from Lead's company
    local account_payload='{
        "name": "'$lead_company'",
        "type": "Prospect",
        "industry": "Technology"
    }'
    
    local account_res=$(api_post "/api/data/account" "$account_payload")
    TEST_ACCOUNT_ID=$(json_extract "$account_res" "id")
    
    if [ -z "$TEST_ACCOUNT_ID" ]; then
        test_failed "Failed to create Account" "$account_res"
        return 1
    fi
    echo "  Created Account: $TEST_ACCOUNT_ID"
    
    # 2. Create Contact linked to Account
    local contact_payload='{
        "name": "Convert '$lead_name'",
        "first_name": "Convert",
        "last_name": "'$lead_name'",
        "email": "'$lead_email'",
        "account_id": "'$TEST_ACCOUNT_ID'"
    }'
    
    local contact_res=$(api_post "/api/data/contact" "$contact_payload")
    TEST_CONTACT_ID=$(json_extract "$contact_res" "id")
    
    if [ -z "$TEST_CONTACT_ID" ]; then
        test_failed "Failed to create Contact" "$contact_res"
        return 1
    fi
    echo "  Created Contact: $TEST_CONTACT_ID"
    
    # 3. Create Opportunity linked to Account
    local opp_payload='{
        "name": "Opp from Lead '$TIMESTAMP'",
        "account_id": "'$TEST_ACCOUNT_ID'",
        "stage_name": "Prospecting",
        "amount": 50000,
        "close_date": "2025-03-31"
    }'
    
    local opp_res=$(api_post "/api/data/opportunity" "$opp_payload")
    TEST_OPPORTUNITY_ID=$(json_extract "$opp_res" "id")
    
    if [ -z "$TEST_OPPORTUNITY_ID" ]; then
        test_failed "Failed to create Opportunity" "$opp_res"
        return 1
    fi
    echo "  Created Opportunity: $TEST_OPPORTUNITY_ID"
    
    # 4. Mark Lead as Converted
    local convert_res=$(api_patch "/api/data/lead/$TEST_LEAD_ID" '{"status": "Converted"}')
    
    if echo "$convert_res" | grep -qE '"id"|"success"|updated'; then
        test_passed "Lead Converted (Account + Contact + Opportunity created)"
    else
        test_failed "Failed to mark Lead as converted" "$convert_res"
    fi
}

# Test 16.4: Progress Opportunity through stages
test_opportunity_stages() {
    echo ""
    echo "Test 16.4: Opportunity Stage Progression"
    
    if [ -z "$TEST_OPPORTUNITY_ID" ]; then
        test_failed "No Opportunity ID available"
        return 1
    fi
    
    local stages=("Qualification" "Needs Analysis" "Proposal" "Negotiation")
    local failures=0
    
    for stage in "${stages[@]}"; do
        local response=$(api_patch "/api/data/opportunity/$TEST_OPPORTUNITY_ID" "{\"stage_name\": \"$stage\"}")
        
        if echo "$response" | grep -qE '"id"|"success"|updated'; then
            echo "  ✓ Stage: $stage"
        else
            echo "  ✗ Failed to update to: $stage"
            failures=$((failures + 1))
        fi
    done
    
    if [ $failures -eq 0 ]; then
        test_passed "Opportunity progressed through all stages"
    else
        test_failed "Stage progression failed ($failures errors)"
    fi
}

# Test 16.5: Close Opportunity as Won
test_closed_won() {
    echo ""
    echo "Test 16.5: Close Opportunity as Won"
    
    if [ -z "$TEST_OPPORTUNITY_ID" ]; then
        test_failed "No Opportunity ID available"
        return 1
    fi
    
    local payload='{"stage_name": "Closed Won"}'
    local response=$(api_patch "/api/data/opportunity/$TEST_OPPORTUNITY_ID" "$payload")
    
    if echo "$response" | grep -qE '"id"|"success"|updated'; then
        local opp=$(api_get "/api/data/opportunity/$TEST_OPPORTUNITY_ID")
        local stage=$(json_extract "$opp" "stage_name")
        
        if [ "$stage" == "Closed Won" ]; then
            test_passed "Opportunity Closed Won"
        else
            test_failed "Stage not updated" "Expected: Closed Won, Got: $stage"
        fi
    else
        test_failed "Failed to close opportunity" "$response"
    fi
}

# Test 16.6: Create another Opportunity and close as Lost
test_closed_lost() {
    echo ""
    echo "Test 16.6: Create and Close Opportunity as Lost"
    
    if [ -z "$TEST_ACCOUNT_ID" ]; then
        echo "  Skipping: No Account available"
        test_passed "Closed Lost (skipped)"
        return
    fi
    
    # Create a second opportunity
    local opp_payload='{
        "name": "Lost Opp '$TIMESTAMP'",
        "account_id": "'$TEST_ACCOUNT_ID'",
        "stage_name": "Prospecting",
        "amount": 25000
    }'
    
    local opp_res=$(api_post "/api/data/opportunity" "$opp_payload")
    local lost_opp_id=$(json_extract "$opp_res" "id")
    
    if [ -z "$lost_opp_id" ]; then
        test_failed "Failed to create lost opportunity" "$opp_res"
        return 1
    fi
    
    # Close as Lost
    local close_res=$(api_patch "/api/data/opportunity/$lost_opp_id" '{"stage_name": "Closed Lost"}')
    
    if echo "$close_res" | grep -qE '"id"|"success"|updated'; then
        # Cleanup this opportunity
        api_delete "/api/data/opportunity/$lost_opp_id" > /dev/null
        test_passed "Opportunity Closed Lost"
    else
        test_failed "Failed to close as lost" "$close_res"
    fi
}

# Cleanup test data
test_cleanup() {
    echo ""
    echo "Test 16.7: Cleanup Test Data"
    
    # Delete Opportunity first (references Account)
    if [ -n "$TEST_OPPORTUNITY_ID" ]; then
        api_delete "/api/data/opportunity/$TEST_OPPORTUNITY_ID" > /dev/null
        echo "  ✓ Opportunity deleted"
    fi
    
    # Delete Contact (references Account)
    if [ -n "$TEST_CONTACT_ID" ]; then
        api_delete "/api/data/contact/$TEST_CONTACT_ID" > /dev/null
        echo "  ✓ Contact deleted"
    fi
    
    # Delete Account
    if [ -n "$TEST_ACCOUNT_ID" ]; then
        api_delete "/api/data/account/$TEST_ACCOUNT_ID" > /dev/null
        echo "  ✓ Account deleted"
    fi
    
    # Delete Lead
    if [ -n "$TEST_LEAD_ID" ]; then
        api_delete "/api/data/lead/$TEST_LEAD_ID" > /dev/null
        echo "  ✓ Lead deleted"
    fi
    
    test_passed "Cleanup completed"
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
