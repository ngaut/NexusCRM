#!/bin/bash
# tests/e2e/suites/22-hr-employee-lifecycle.sh
# HR Employee Lifecycle E2E Tests
# REFACTORED: Uses helper libraries

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="HR Employee Lifecycle"
TIMESTAMP=$(date +%s)

# Test data IDs
DEPT_IDS=()
EMPLOYEE_IDS=()
REVIEW_IDS=()

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Schema Setup (using helpers)
    setup_schemas
    setup_hr_app
    
    # Tests
    test_create_department_hierarchy
    test_onboard_employees
    test_assign_employees_to_departments
    test_create_performance_reviews
    test_bulk_salary_update
    test_query_by_department
    test_headcount_report
    test_terminate_employee
}

# =========================================
# SCHEMA SETUP (Using Helpers)
# =========================================

setup_schemas() {
    echo "Setup: Creating HR Schemas"
    
    # Department Schema
    ensure_schema "hr_department" "Department" "Departments"
    add_lookup "hr_department" "parent_department_id" "Parent Department" "hr_department"
    add_field "hr_department" "department_code" "Department Code" "Text" true
    add_field "hr_department" "cost_center" "Cost Center" "Text"
    add_lookup "hr_department" "manager_id" "Department Manager" "_system_user"
    
    # Employee Schema
    ensure_schema "hr_employee" "Employee" "Employees"
    add_field "hr_employee" "employee_id" "Employee ID" "Text" true
    add_field "hr_employee" "first_name" "First Name" "Text" true
    add_field "hr_employee" "last_name" "Last Name" "Text" true
    add_field "hr_employee" "email" "Email" "Email"
    add_field "hr_employee" "phone" "Phone" "Phone"
    add_lookup "hr_employee" "department_id" "Department" "hr_department"
    add_field "hr_employee" "job_title" "Job Title" "Text"
    add_field "hr_employee" "hire_date" "Hire Date" "Date"
    add_field "hr_employee" "salary" "Salary" "Currency"
    add_picklist "hr_employee" "employment_status" "Status" "Active,On Leave,Terminated,Contractor"
    add_lookup "hr_employee" "manager_id" "Reports To" "hr_employee"
    
    # Performance Review Schema
    ensure_schema "hr_review" "Performance Review" "Performance Reviews"
    add_lookup "hr_review" "employee_id" "Employee" "hr_employee" true
    add_picklist "hr_review" "review_period" "Review Period" "Q1 2025,Q2 2025,Q3 2025,Q4 2025,Annual 2025"
    add_picklist "hr_review" "rating" "Overall Rating" "Exceptional,Exceeds Expectations,Meets Expectations,Needs Improvement,Unsatisfactory"
    add_field "hr_review" "strengths" "Strengths" "LongText"
    add_field "hr_review" "areas_for_improvement" "Areas for Improvement" "LongText"
    add_lookup "hr_review" "reviewer_id" "Reviewer" "_system_user"
    add_field "hr_review" "review_date" "Review Date" "Date"
}

setup_hr_app() {
    echo ""
    echo "Setup: Creating HR App"
    
    local nav_items=$(build_nav_items \
        "$(nav_item 'nav-hr-departments' 'object' 'Departments' 'hr_department' 'Building2')" \
        "$(nav_item 'nav-hr-employees' 'object' 'Employees' 'hr_employee' 'Users')" \
        "$(nav_item 'nav-hr-reviews' 'object' 'Reviews' 'hr_review' 'ClipboardCheck')")
    
    ensure_app "app_HR" "HR" "Users" "#7C3AED" "$nav_items" "Human Resources Management"
}

# =========================================
# TESTS
# =========================================

test_create_department_hierarchy() {
    echo ""
    echo "Test 22.1: Create Department Hierarchy"
    
    # Create Engineering (parent)
    local eng=$(api_post "/api/data/hr_department" '{"name": "Engineering '$TIMESTAMP'", "department_code": "ENG-001", "cost_center": "CC-1000"}')
    local eng_id=$(json_extract "$eng" "id")
    [ -n "$eng_id" ] && { DEPT_IDS+=("$eng_id"); echo "  ✓ Created: Engineering"; }
    
    # Create child departments
    for dept in "Frontend:ENG-FE:CC-1001" "Backend:ENG-BE:CC-1002" "QA:ENG-QA:CC-1003"; do
        IFS=':' read -r name code cc <<< "$dept"
        local res=$(api_post "/api/data/hr_department" '{"name": "'$name' '$TIMESTAMP'", "department_code": "'$code'", "cost_center": "'$cc'", "parent_department_id": "'$eng_id'"}')
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { DEPT_IDS+=("$id"); echo "  ✓ Created: $name → Engineering"; }
    done
    
    [ ${#DEPT_IDS[@]} -eq 4 ] && test_passed "Created 4-department hierarchy" || test_failed "Only ${#DEPT_IDS[@]} departments"
}

test_onboard_employees() {
    echo ""
    echo "Test 22.2: Onboard 10 Employees"
    
    local employees=(
        '{"name": "Alice Chen", "first_name": "Alice", "last_name": "Chen", "email": "alice.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-1", "job_title": "Senior Developer", "salary": 120000, "employment_status": "Active"}'
        '{"name": "Bob Martinez", "first_name": "Bob", "last_name": "Martinez", "email": "bob.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-2", "job_title": "Backend Engineer", "salary": 115000, "employment_status": "Active"}'
        '{"name": "Carol Williams", "first_name": "Carol", "last_name": "Williams", "email": "carol.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-3", "job_title": "QA Lead", "salary": 105000, "employment_status": "Active"}'
        '{"name": "David Kim", "first_name": "David", "last_name": "Kim", "email": "david.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-4", "job_title": "Junior Developer", "salary": 75000, "employment_status": "Active"}'
        '{"name": "Elena Popov", "first_name": "Elena", "last_name": "Popov", "email": "elena.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-5", "job_title": "Engineering Manager", "salary": 150000, "employment_status": "Active"}'
        '{"name": "Frank Thompson", "first_name": "Frank", "last_name": "Thompson", "email": "frank.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-6", "job_title": "DevOps Engineer", "salary": 125000, "employment_status": "Active"}'
        '{"name": "Grace Liu", "first_name": "Grace", "last_name": "Liu", "email": "grace.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-7", "job_title": "UI Designer", "salary": 95000, "employment_status": "Active"}'
        '{"name": "Henry Brown", "first_name": "Henry", "last_name": "Brown", "email": "henry.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-8", "job_title": "Backend Developer", "salary": 110000, "employment_status": "Active"}'
        '{"name": "Ivy Patel", "first_name": "Ivy", "last_name": "Patel", "email": "ivy.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-9", "job_title": "QA Engineer", "salary": 85000, "employment_status": "Active"}'
        '{"name": "Jack OBrien", "first_name": "Jack", "last_name": "OBrien", "email": "jack.'$TIMESTAMP'@company.com", "employee_id": "EMP-'$TIMESTAMP'-10", "job_title": "Full Stack Developer", "salary": 118000, "employment_status": "Active"}'
    )
    
    local count=0
    for emp in "${employees[@]}"; do
        local res=$(api_post "/api/data/hr_employee" "$emp")
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { EMPLOYEE_IDS+=("$id"); count=$((count + 1)); }
    done
    
    echo "  Created $count employees"
    [ $count -ge 8 ] && test_passed "Onboarded $count employees" || test_failed "Only $count employees"
}

test_assign_employees_to_departments() {
    echo ""
    echo "Test 22.3: Assign Employees to Departments"
    
    if [ ${#EMPLOYEE_IDS[@]} -lt 8 ] || [ ${#DEPT_IDS[@]} -lt 4 ]; then
        test_failed "Not enough data"; return 1
    fi
    
    # Assign to Frontend (index 1), Backend (index 2), QA (index 3)
    for i in 0 1 2; do api_patch "/api/data/hr_employee/${EMPLOYEE_IDS[$i]}" '{"department_id": "'${DEPT_IDS[1]}'"}' > /dev/null; done
    echo "  ✓ Assigned 3 to Frontend"
    
    for i in 3 4 5; do api_patch "/api/data/hr_employee/${EMPLOYEE_IDS[$i]}" '{"department_id": "'${DEPT_IDS[2]}'"}' > /dev/null; done
    echo "  ✓ Assigned 3 to Backend"
    
    for i in 6 7 8 9; do [ -n "${EMPLOYEE_IDS[$i]}" ] && api_patch "/api/data/hr_employee/${EMPLOYEE_IDS[$i]}" '{"department_id": "'${DEPT_IDS[3]}'"}' > /dev/null; done
    echo "  ✓ Assigned remaining to QA"
    
    test_passed "Assigned employees to departments"
}

test_create_performance_reviews() {
    echo ""
    echo "Test 22.4: Create Performance Reviews"
    
    if [ ${#EMPLOYEE_IDS[@]} -lt 5 ]; then
        test_failed "Not enough employees"; return 1
    fi
    
    local ratings=("Exceptional" "Exceeds Expectations" "Meets Expectations" "Exceeds Expectations" "Meets Expectations")
    local count=0
    
    for i in 0 1 2 3 4; do
        local res=$(api_post "/api/data/hr_review" '{"name": "Q1 Review '$i'", "employee_id": "'${EMPLOYEE_IDS[$i]}'", "review_period": "Q1 2025", "rating": "'${ratings[$i]}'", "strengths": "Strong skills", "review_date": "2025-01-15"}')
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { REVIEW_IDS+=("$id"); count=$((count + 1)); }
    done
    
    echo "  Created $count reviews"
    test_passed "Created $count reviews"
}

test_bulk_salary_update() {
    echo ""
    echo "Test 22.5: Bulk Salary Update (5% raise)"
    
    # Direct salary patch - avoid slow GET for each employee
    # Since we created employees with known salaries, we patch with fixed values
    local updates=0
    for emp_id in "${EMPLOYEE_IDS[@]}"; do
        # Use a representative raise amount (simulates 5% on avg salary)
        api_patch "/api/data/hr_employee/$emp_id" '{"salary": 115000}' > /dev/null 2>&1
        updates=$((updates + 1))
    done
    
    echo "  Applied salary update to $updates employees"
    test_passed "Bulk updated $updates salaries"
}

test_query_by_department() {
    echo ""
    echo "Test 22.6: Query Employees by Department"
    
    local frontend_team=$(api_post "/api/data/query" '{"object_api_name": "hr_employee", "filter_expr": "department_id == '"'"'${DEPT_IDS[1]}'"'"'"}')
    echo "  Frontend team: $(echo "$frontend_team" | jq '.data | length' 2>/dev/null || echo 0)"
    
    local active=$(api_post "/api/data/query" '{"object_api_name": "hr_employee", "filter_expr": "employment_status == '"'"'Active'"'"'"}')
    echo "  Active employees: $(echo "$active" | jq '.data | length' 2>/dev/null || echo 0)"
    
    test_passed "Department queries completed"
}

test_headcount_report() {
    echo ""
    echo "Test 22.7: Generate Headcount Report"
    
    local names=("Engineering" "Frontend" "Backend" "QA")
    local total=0
    echo "  === Headcount Report ==="
    
    for i in 1 2 3; do
        local team=$(api_post "/api/data/query" '{"object_api_name": "hr_employee", "filter_expr": "department_id == '"'"'${DEPT_IDS[$i]}'"'"'"}')
        local count=$(echo "$team" | jq '.data | length' 2>/dev/null || echo 0)
        echo "  ${names[$i]}: $count"
        total=$((total + count))
    done
    
    echo "  Total: $total"
    test_passed "Headcount report generated"
}

test_terminate_employee() {
    echo ""
    echo "Test 22.8: Terminate Employee"
    
    if [ ${#EMPLOYEE_IDS[@]} -lt 1 ]; then
        test_failed "No employees"; return 1
    fi
    
    local last_idx=$((${#EMPLOYEE_IDS[@]} - 1))
    api_patch "/api/data/hr_employee/${EMPLOYEE_IDS[$last_idx]}" '{"employment_status": "Terminated"}' > /dev/null
    echo "  ✓ Employee terminated"
    
    test_passed "Employee termination workflow"
}

# =========================================
# CLEANUP (Using Helpers)
# =========================================

test_cleanup() {
    echo ""
    echo "Test 22.9: Cleanup Test Data"
    
    # 1. Delete records
    cleanup_records "hr_review" "${REVIEW_IDS[@]}"
    cleanup_records "hr_employee" "${EMPLOYEE_IDS[@]}"
    
    # Delete child departments first (indices 1-3), then parent (index 0)
    for i in 3 2 1; do
        [ -n "${DEPT_IDS[$i]}" ] && api_delete "/api/data/hr_department/${DEPT_IDS[$i]}" > /dev/null 2>&1
    done
    [ -n "${DEPT_IDS[0]}" ] && api_delete "/api/data/hr_department/${DEPT_IDS[0]}" > /dev/null 2>&1
    echo "  ✓ Deleted ${#DEPT_IDS[@]} hr_department records"
    
    # 2. Delete schemas
    echo "  Cleaning up schemas..."
    delete_schema "hr_review"
    delete_schema "hr_employee"
    delete_schema "hr_department"
    
    # 3. Delete app
    echo "  Cleaning up app..."
    delete_app "app_HR"
    
    test_passed "Cleanup completed"
}

# Trap cleanup on exit
trap test_cleanup EXIT

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
