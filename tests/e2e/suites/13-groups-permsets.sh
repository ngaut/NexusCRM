#!/bin/bash
# tests/e2e/suites/13-groups-permsets.sh
# User Groups (Queues) and Permission Sets E2E Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Groups & Permission Sets"

# Test data IDs (will be set during tests)
TEST_GROUP_ID=""
TEST_PERMSET_ID=""
TEST_USER_ID=""

run_suite() {
    section_header "$SUITE_NAME"
    
    # Ensure we're logged in
    if [ -z "$TOKEN" ]; then
        if ! api_login "$TEST_EMAIL" "$TEST_PASSWORD"; then
            echo -e "${RED}CRITICAL: Cannot login for tests${NC}"
            return 1
        fi
    fi
    
    test_schema_exists
    test_create_group
    test_create_permission_set
    test_assign_permission_set
    test_group_membership
    test_cleanup
}

# Test that the new system tables exist in schema
test_schema_exists() {
    echo "Test 13.1: Verify New System Tables Exist"
    
    # Check _System_Group schema
    local response=$(api_get "/api/metadata/objects/_system_group")
    if echo "$response" | grep -q '"api_name"'; then
        test_passed "_System_Group schema exists"
    else
        test_failed "_System_Group schema not found" "$response"
    fi
    
    echo ""
    # Check _System_PermissionSet schema
    response=$(api_get "/api/metadata/objects/_system_permissionset")
    if echo "$response" | grep -q '"api_name"'; then
        test_passed "_System_PermissionSet schema exists"
    else
        test_failed "_System_PermissionSet schema not found" "$response"
    fi
    
    echo ""
    # Check _System_GroupMember schema
    response=$(api_get "/api/metadata/objects/_system_groupmember")
    if echo "$response" | grep -q '"api_name"'; then
        test_passed "_System_GroupMember schema exists"
    else
        test_failed "_System_GroupMember schema not found" "$response"
    fi
    
    echo ""
    # Check _System_PermissionSetAssignment schema
    response=$(api_get "/api/metadata/objects/_system_permissionsetassignment")
    if echo "$response" | grep -q '"api_name"'; then
        test_passed "_System_PermissionSetAssignment schema exists"
    else
        test_failed "_System_PermissionSetAssignment schema not found" "$response"
    fi
}

# Test creating a Group (Queue)
test_create_group() {
    echo ""
    echo "Test 13.2: Create a Group (Queue)"
    
    local group_name="e2e_test_queue_$(date +%s)"
    local payload='{
        "name": "'$group_name'",
        "label": "E2E Test Queue",
        "type": "Queue",
        "email": "testqueue@example.com"
    }'
    
    local response=$(api_post "/api/data/_system_group" "$payload")
    
    if echo "$response" | grep -q '"id"'; then
        TEST_GROUP_ID=$(echo "$response" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')
        test_passed "Group created successfully"
        echo -e "  ${BLUE}Group ID:${NC} $TEST_GROUP_ID"
    else
        test_failed "Create Group" "$response"
    fi
}

# Test creating a Permission Set
test_create_permission_set() {
    echo ""
    echo "Test 13.3: Create a Permission Set"
    
    local permset_name="e2e_test_permset_$(date +%s)"
    local payload='{
        "name": "'$permset_name'",
        "label": "E2E Test Permission Set",
        "description": "Permission set created for E2E testing",
        "is_active": true
    }'
    
    local response=$(api_post "/api/data/_system_permissionset" "$payload")
    
    if echo "$response" | grep -q '"id"'; then
        TEST_PERMSET_ID=$(echo "$response" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')
        test_passed "Permission Set created successfully"
        echo -e "  ${BLUE}Permission Set ID:${NC} $TEST_PERMSET_ID"
    else
        test_failed "Create Permission Set" "$response"
    fi
}

# Test assigning Permission Set to current user
test_assign_permission_set() {
    echo ""
    echo "Test 13.4: Assign Permission Set to User"
    
    if [ -z "$TEST_PERMSET_ID" ] || [ -z "$USER_ID" ]; then
        test_failed "Cannot test assignment - missing IDs"
        return
    fi
    
    local payload='{
        "assignee_id": "'$USER_ID'",
        "permission_set_id": "'$TEST_PERMSET_ID'"
    }'
    
    local response=$(api_post "/api/data/_system_permissionsetassignment" "$payload")
    
    if echo "$response" | grep -q '"id"'; then
        local assignment_id=$(echo "$response" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')
        test_passed "Permission Set assigned to user"
        echo -e "  ${BLUE}Assignment ID:${NC} $assignment_id"
        
        # Verify assignment by querying
        echo ""
        echo "Test 13.4b: Verify Assignment via Query"
        local query_response=$(api_post "/api/data/query" '{
            "object_api_name": "_system_permissionsetassignment",
            "filters": [
                {"field": "assignee_id", "operator": "=", "value": "'$USER_ID'"}
            ]
        }')
        
        if echo "$query_response" | grep -q "$TEST_PERMSET_ID"; then
            test_passed "Assignment verified via query"
        else
            test_failed "Assignment verification" "$query_response"
        fi
    else
        test_failed "Assign Permission Set" "$response"
    fi
}

# Test adding user to Group
test_group_membership() {
    echo ""
    echo "Test 13.5: Add User to Group"
    
    if [ -z "$TEST_GROUP_ID" ] || [ -z "$USER_ID" ]; then
        test_failed "Cannot test group membership - missing IDs"
        return
    fi
    
    local payload='{
        "group_id": "'$TEST_GROUP_ID'",
        "user_id": "'$USER_ID'"
    }'
    
    local response=$(api_post "/api/data/_system_groupmember" "$payload")
    
    if echo "$response" | grep -q '"id"'; then
        local member_id=$(echo "$response" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')
        test_passed "User added to Group"
        echo -e "  ${BLUE}Membership ID:${NC} $member_id"
        
        # Verify membership by querying
        echo ""
        echo "Test 13.5b: Verify Group Membership via Query"
        local query_response=$(api_post "/api/data/query" '{
            "object_api_name": "_system_groupmember",
            "filters": [
                {"field": "group_id", "operator": "=", "value": "'$TEST_GROUP_ID'"}
            ]
        }')
        
        if echo "$query_response" | grep -q "$USER_ID"; then
            test_passed "Group membership verified via query"
        else
            test_failed "Group membership verification" "$query_response"
        fi
    else
        test_failed "Add User to Group" "$response"
    fi
}

# Cleanup test data
test_cleanup() {
    echo ""
    echo "Test 13.6: Cleanup Test Data"
    
    local cleanup_success=true
    
    # Delete group member (will cascade when group is deleted, but let's be explicit)
    if [ -n "$TEST_GROUP_ID" ]; then
        local response=$(api_delete "/api/data/_system_group/$TEST_GROUP_ID")
        if echo "$response" | grep -qE '"success"|deleted|200'; then
            echo -e "  ${GREEN}✓${NC} Group deleted"
        else
            echo -e "  ${YELLOW}⚠${NC} Group deletion response: $response"
        fi
    fi
    
    # Delete permission set (will cascade assignments)
    if [ -n "$TEST_PERMSET_ID" ]; then
        local response=$(api_delete "/api/data/_system_permissionset/$TEST_PERMSET_ID")
        if echo "$response" | grep -qE '"success"|deleted|200'; then
            echo -e "  ${GREEN}✓${NC} Permission Set deleted"
        else
            echo -e "  ${YELLOW}⚠${NC} Permission Set deletion response: $response"
        fi
    fi
    
    test_passed "Cleanup completed"
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
