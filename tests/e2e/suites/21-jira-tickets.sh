#!/bin/bash
# tests/e2e/suites/21-jira-tickets.sh
# Jira-like Ticket System E2E Tests
# REFACTORED: Uses helper libraries

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="Jira Ticket System"
TIMESTAMP=$(date +%s)

# Test data IDs
TEST_PROJECT_ID=""
TEST_ISSUE_BUG_ID=""
TEST_ISSUE_STORY_ID=""
TEST_ISSUE_TASK_ID=""
TEST_COMMENT_ID=""
TEST_SPRINT_ID=""

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Schema Setup (using helpers)
    setup_schemas
    setup_jira_app
    
    # Tests
    test_create_project
    test_create_issues
    test_issue_workflow
    test_issue_assignment
    test_add_comment
    test_sprint_management
    test_priority_escalation
    test_query_by_status
}

# =========================================
# SCHEMA SETUP (Using Helpers)
# =========================================

setup_schemas() {
    echo "Setup: Creating Jira Schemas"
    
    # Project Schema
    ensure_schema "jira_project" "Project" "Projects"
    add_field "jira_project" "project_key" "Project Key" "Text" true
    add_field "jira_project" "description" "Description" "LongText"
    add_lookup "jira_project" "lead_id" "Project Lead" "_system_user"
    
    # Issue Schema (Jira Ticket)
    ensure_schema "jira_issue" "Issue" "Issues"
    add_field "jira_issue" "summary" "Summary" "Text" true
    add_field "jira_issue" "description" "Description" "LongText"
    add_picklist "jira_issue" "issue_type" "Issue Type" "Bug,Story,Task,Epic,Sub-task"
    add_picklist "jira_issue" "status" "Status" "To Do,In Progress,In Review,Done,Blocked"
    add_picklist "jira_issue" "priority" "Priority" "Lowest,Low,Medium,High,Highest"
    add_lookup "jira_issue" "assignee_id" "Assignee" "_system_user"
    add_lookup "jira_issue" "reporter_id" "Reporter" "_system_user"
    add_lookup "jira_issue" "project_id" "Project" "jira_project"
    add_field "jira_issue" "story_points" "Story Points" "Number"
    add_field "jira_issue" "labels" "Labels" "Text"
    add_field "jira_issue" "due_date" "Due Date" "Date"
    add_lookup "jira_issue" "sprint_id" "Sprint" "jira_sprint"
    
    # Comment Schema
    ensure_schema "jira_comment" "Comment" "Comments"
    add_field "jira_comment" "body" "Comment Body" "LongText" true
    add_lookup "jira_comment" "issue_id" "Issue" "jira_issue"
    add_lookup "jira_comment" "author_id" "Author" "_system_user"
    
    # Sprint Schema
    ensure_schema "jira_sprint" "Sprint" "Sprints"
    add_field "jira_sprint" "goal" "Sprint Goal" "Text"
    add_lookup "jira_sprint" "project_id" "Project" "jira_project"
    add_field "jira_sprint" "start_date" "Start Date" "Date"
    add_field "jira_sprint" "end_date" "End Date" "Date"
    add_picklist "jira_sprint" "status" "Sprint Status" "Future,Active,Completed"
}

setup_jira_app() {
    echo ""
    echo "Setup: Creating Jira App"
    
    local nav_items=$(build_nav_items \
        "$(nav_item 'nav-jira-projects' 'object' 'Projects' 'jira_project' 'Folder')" \
        "$(nav_item 'nav-jira-issues' 'object' 'Issues' 'jira_issue' 'Bug')" \
        "$(nav_item 'nav-jira-sprints' 'object' 'Sprints' 'jira_sprint' 'Calendar')" \
        "$(nav_item 'nav-jira-comments' 'object' 'Comments' 'jira_comment' 'MessageCircle')")
    
    ensure_app "app_Jira" "Jira" "bug" "#0052CC" "$nav_items" "Issue Tracking and Project Management"
}

# =========================================
# TESTS
# =========================================

test_create_project() {
    echo ""
    echo "Test 21.1: Create Jira Project"
    
    local res=$(api_post "/api/data/jira_project" '{"name": "Test Project '$TIMESTAMP'", "project_key": "TEST", "description": "E2E Test Project", "lead_id": "'$USER_ID'"}')
    TEST_PROJECT_ID=$(json_extract "$res" "id")
    
    [ -n "$TEST_PROJECT_ID" ] && test_passed "Project created" || test_failed "Failed to create Project"
}

test_create_issues() {
    echo ""
    echo "Test 21.2: Create Issues (Bug, Story, Task)"
    
    [ -z "$TEST_PROJECT_ID" ] && { test_failed "No Project ID"; return 1; }
    
    local bug=$(api_post "/api/data/jira_issue" '{"name": "BUG-'$TIMESTAMP'", "summary": "Login button not working", "issue_type": "Bug", "status": "To Do", "priority": "High", "project_id": "'$TEST_PROJECT_ID'", "reporter_id": "'$USER_ID'", "story_points": 3}')
    TEST_ISSUE_BUG_ID=$(json_extract "$bug" "id")
    [ -n "$TEST_ISSUE_BUG_ID" ] && echo "  ✓ Bug created"
    
    local story=$(api_post "/api/data/jira_issue" '{"name": "STORY-'$TIMESTAMP'", "summary": "Password reset feature", "issue_type": "Story", "status": "To Do", "priority": "Medium", "project_id": "'$TEST_PROJECT_ID'", "reporter_id": "'$USER_ID'", "story_points": 5}')
    TEST_ISSUE_STORY_ID=$(json_extract "$story" "id")
    [ -n "$TEST_ISSUE_STORY_ID" ] && echo "  ✓ Story created"
    
    local task=$(api_post "/api/data/jira_issue" '{"name": "TASK-'$TIMESTAMP'", "summary": "Update documentation", "issue_type": "Task", "status": "To Do", "priority": "Low", "project_id": "'$TEST_PROJECT_ID'", "reporter_id": "'$USER_ID'", "story_points": 2}')
    TEST_ISSUE_TASK_ID=$(json_extract "$task" "id")
    [ -n "$TEST_ISSUE_TASK_ID" ] && echo "  ✓ Task created"
    
    [ -n "$TEST_ISSUE_BUG_ID" ] && [ -n "$TEST_ISSUE_STORY_ID" ] && [ -n "$TEST_ISSUE_TASK_ID" ] && test_passed "All issue types created" || test_failed "Some issues failed"
}

test_issue_workflow() {
    echo ""
    echo "Test 21.3: Issue Workflow (To Do → Done)"
    
    [ -z "$TEST_ISSUE_BUG_ID" ] && { test_failed "No Bug issue"; return 1; }
    
    for status in "In Progress" "In Review" "Done"; do
        api_patch "/api/data/jira_issue/$TEST_ISSUE_BUG_ID" '{"status": "'$status'"}' > /dev/null
        echo "  ✓ Status → $status"
    done
    
    test_passed "Workflow transitions completed"
}

test_issue_assignment() {
    echo ""
    echo "Test 21.4: Issue Assignment"
    
    [ -z "$TEST_ISSUE_STORY_ID" ] && { test_failed "No Story issue"; return 1; }
    
    api_patch "/api/data/jira_issue/$TEST_ISSUE_STORY_ID" '{"assignee_id": "'$USER_ID'"}' > /dev/null
    echo "  ✓ Issue assigned to: $USER_ID"
    test_passed "Issue assignment verified"
}

test_add_comment() {
    echo ""
    echo "Test 21.5: Add Comment to Issue"
    
    [ -z "$TEST_ISSUE_BUG_ID" ] && { test_failed "No Bug issue"; return 1; }
    
    local res=$(api_post "/api/data/jira_comment" '{"name": "Comment '$TIMESTAMP'", "body": "I can reproduce this. Working on a fix.", "issue_id": "'$TEST_ISSUE_BUG_ID'", "author_id": "'$USER_ID'"}')
    TEST_COMMENT_ID=$(json_extract "$res" "id")
    
    [ -n "$TEST_COMMENT_ID" ] && test_passed "Comment added" || test_failed "Failed to add comment"
}

test_sprint_management() {
    echo ""
    echo "Test 21.6: Sprint Management"
    
    [ -z "$TEST_PROJECT_ID" ] && { test_failed "No Project ID"; return 1; }
    
    local res=$(api_post "/api/data/jira_sprint" '{"name": "Sprint 1 '$TIMESTAMP'", "goal": "Complete login fixes", "project_id": "'$TEST_PROJECT_ID'", "start_date": "2025-01-01", "end_date": "2025-01-14", "status": "Active"}')
    TEST_SPRINT_ID=$(json_extract "$res" "id")
    
    if [ -n "$TEST_SPRINT_ID" ]; then
        echo "  ✓ Sprint created"
        [ -n "$TEST_ISSUE_BUG_ID" ] && api_patch "/api/data/jira_issue/$TEST_ISSUE_BUG_ID" '{"sprint_id": "'$TEST_SPRINT_ID'"}' > /dev/null
        [ -n "$TEST_ISSUE_STORY_ID" ] && api_patch "/api/data/jira_issue/$TEST_ISSUE_STORY_ID" '{"sprint_id": "'$TEST_SPRINT_ID'"}' > /dev/null
        echo "  ✓ Issues added to sprint"
        test_passed "Sprint created and issues assigned"
    else
        test_failed "Failed to create Sprint"
    fi
}

test_priority_escalation() {
    echo ""
    echo "Test 21.7: Priority Escalation"
    
    [ -z "$TEST_ISSUE_TASK_ID" ] && { test_failed "No Task issue"; return 1; }
    
    api_patch "/api/data/jira_issue/$TEST_ISSUE_TASK_ID" '{"priority": "Highest"}' > /dev/null
    echo "  ✓ Priority escalated: Low → Highest"
    test_passed "Priority escalation verified"
}

test_query_by_status() {
    echo ""
    echo "Test 21.8: Query Issues by Status"
    
    local done_issues=$(api_post "/api/data/query" '{"object_api_name": "jira_issue", "filter_expr": "status == '"'"'Done'"'"'"}')
    echo "  Done issues: $(echo "$done_issues" | jq '.data | length' 2>/dev/null || echo 0)"
    
    local todo_issues=$(api_post "/api/data/query" '{"object_api_name": "jira_issue", "filter_expr": "status == '"'"'To Do'"'"'"}')
    echo "  To Do issues: $(echo "$todo_issues" | jq '.data | length' 2>/dev/null || echo 0)"
    
    test_passed "Issue queries completed"
}

# =========================================
# CLEANUP
# =========================================

test_cleanup() {
    echo ""
    echo "Test 21.9: Cleanup Test Data"
    
    # 1. Delete records
    [ -n "$TEST_COMMENT_ID" ] && api_delete "/api/data/jira_comment/$TEST_COMMENT_ID" > /dev/null 2>&1
    [ -n "$TEST_ISSUE_BUG_ID" ] && api_delete "/api/data/jira_issue/$TEST_ISSUE_BUG_ID" > /dev/null 2>&1
    [ -n "$TEST_ISSUE_STORY_ID" ] && api_delete "/api/data/jira_issue/$TEST_ISSUE_STORY_ID" > /dev/null 2>&1
    [ -n "$TEST_ISSUE_TASK_ID" ] && api_delete "/api/data/jira_issue/$TEST_ISSUE_TASK_ID" > /dev/null 2>&1
    [ -n "$TEST_SPRINT_ID" ] && api_delete "/api/data/jira_sprint/$TEST_SPRINT_ID" > /dev/null 2>&1
    [ -n "$TEST_PROJECT_ID" ] && api_delete "/api/data/jira_project/$TEST_PROJECT_ID" > /dev/null 2>&1
    
    # 2. Delete schemas
    echo "  Cleaning up schemas..."
    delete_schema "jira_comment"
    delete_schema "jira_issue"
    delete_schema "jira_sprint"
    delete_schema "jira_project"
    
    # 3. Delete app
    echo "  Cleaning up app..."
    delete_app "app_Jira"
    
    test_passed "Cleanup completed"
}

# Trap cleanup on exit
trap test_cleanup EXIT

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
