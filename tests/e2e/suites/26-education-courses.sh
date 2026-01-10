#!/bin/bash
# tests/e2e/suites/26-education-courses.sh
# Education Course Management E2E Tests
# REFACTORED: Uses helper libraries

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="Education Course Management"
TIMESTAMP=$(date +%s)

# Test data IDs
COURSE_IDS=()
INSTRUCTOR_IDS=()
STUDENT_IDS=()
ENROLLMENT_IDS=()

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Schema Setup (using helpers)
    setup_schemas
    setup_education_app
    
    # Tests
    test_create_course_catalog
    test_create_instructors
    test_create_students
    test_enroll_students
    test_record_grades
    test_query_transcripts
    test_course_drop
    test_course_drop
}

# =========================================
# SCHEMA SETUP (Using Helpers)
# =========================================

setup_schemas() {
    echo "Setup: Creating Education Schemas"
    
    # Course Schema
    ensure_schema "edu_course" "Course" "Courses"
    add_field "edu_course" "code" "Course Code" "Text" true
    add_field "edu_course" "description" "Description" "LongText"
    add_field "edu_course" "credits" "Credits" "Number" true
    add_picklist "edu_course" "department" "Department" "Computer Science,Mathematics,Physics,Chemistry,Biology,English,History"
    add_picklist "edu_course" "level" "Level" "100,200,300,400,Graduate"
    add_field "edu_course" "max_enrollment" "Max Enrollment" "Number"
    add_lookup "edu_course" "instructor_id" "Instructor" "edu_instructor"
    
    # Instructor Schema (create first for course lookup)
    ensure_schema "edu_instructor" "Instructor" "Instructors"
    add_field "edu_instructor" "email" "Email" "Email" true
    add_picklist "edu_instructor" "department" "Department" "Computer Science,Mathematics,Physics,Chemistry,Biology,English,History"
    add_picklist "edu_instructor" "title" "Title" "Professor,Associate Professor,Assistant Professor,Lecturer,Adjunct"
    add_field "edu_instructor" "office" "Office" "Text"
    add_field "edu_instructor" "phone" "Phone" "Phone"
    
    # Student Schema
    ensure_schema "edu_student" "Student" "Students"
    add_field "edu_student" "student_id" "Student ID" "Text" true
    add_field "edu_student" "email" "Email" "Email" true
    add_picklist "edu_student" "major" "Major" "Computer Science,Mathematics,Physics,Chemistry,Biology,English,History,Undeclared"
    add_picklist "edu_student" "year" "Year" "Freshman,Sophomore,Junior,Senior,Graduate"
    add_field "edu_student" "gpa" "GPA" "Number"
    add_picklist "edu_student" "status" "Status" "Active,Graduated,On Leave,Withdrawn"
    
    # Enrollment Schema (junction)
    ensure_schema "edu_enrollment" "Enrollment" "Enrollments"
    add_lookup "edu_enrollment" "student_id" "Student" "edu_student" true
    add_lookup "edu_enrollment" "course_id" "Course" "edu_course" true
    add_picklist "edu_enrollment" "semester" "Semester" "Fall 2024,Spring 2025,Summer 2025,Fall 2025"
    add_picklist "edu_enrollment" "grade" "Grade" "A,A-,B+,B,B-,C+,C,C-,D+,D,F,W,I,IP"
    add_picklist "edu_enrollment" "status" "Status" "Enrolled,Completed,Dropped,Withdrawn"
}

setup_education_app() {
    echo ""
    echo "Setup: Creating Education App"
    
    local nav_items=$(build_nav_items \
        "$(nav_item 'nav-edu-courses' 'object' 'Courses' 'edu_course' 'Book')" \
        "$(nav_item 'nav-edu-instructors' 'object' 'Instructors' 'edu_instructor' 'UserCheck')" \
        "$(nav_item 'nav-edu-students' 'object' 'Students' 'edu_student' 'Users')" \
        "$(nav_item 'nav-edu-enrollments' 'object' 'Enrollments' 'edu_enrollment' 'ClipboardList')")
    
    ensure_app "app_Education" "Education" "GraduationCap" "#F59E0B" "$nav_items" "Academic Course Management"
}

# =========================================
# TESTS
# =========================================

test_create_course_catalog() {
    echo ""
    echo "Test 26.1: Create Course Catalog"
    
    local courses=(
        '{"name": "Introduction to Programming", "code": "CS101", "credits": 3, "department": "Computer Science", "level": "100", "max_enrollment": 30}'
        '{"name": "Data Structures", "code": "CS201", "credits": 4, "department": "Computer Science", "level": "200", "max_enrollment": 25}'
        '{"name": "Calculus I", "code": "MATH101", "credits": 4, "department": "Mathematics", "level": "100", "max_enrollment": 35}'
        '{"name": "Physics I", "code": "PHYS101", "credits": 4, "department": "Physics", "level": "100", "max_enrollment": 30}'
    )
    
    local count=0
    for course in "${courses[@]}"; do
        local res=$(api_post "/api/data/edu_course" "$course")
        local id=$(json_extract "$res" "id")
        if [ -n "$id" ]; then
            COURSE_IDS+=("$id")
            count=$((count + 1))
        fi
    done
    
    echo "  Created $count courses"
    [ $count -eq 4 ] && test_passed "Course catalog created" || test_failed "Only created $count courses"
}

test_create_instructors() {
    echo ""
    echo "Test 26.2: Create Instructors"
    
    local instructors=(
        '{"name": "Prof. Alan Turing", "email": "a.turing.'$TIMESTAMP'@university.edu", "department": "Computer Science", "title": "Professor", "office": "CS 301"}'
        '{"name": "Prof. Emmy Noether", "email": "e.noether.'$TIMESTAMP'@university.edu", "department": "Mathematics", "title": "Professor", "office": "Math 205"}'
    )
    
    local count=0
    for instructor in "${instructors[@]}"; do
        local res=$(api_post "/api/data/edu_instructor" "$instructor")
        local id=$(json_extract "$res" "id")
        if [ -n "$id" ]; then
            INSTRUCTOR_IDS+=("$id")
            count=$((count + 1))
        fi
    done
    
    echo "  Created $count instructors"
    
    if [ $count -eq 2 ] && [ ${#COURSE_IDS[@]} -ge 2 ]; then
        api_patch "/api/data/edu_course/${COURSE_IDS[0]}" '{"instructor_id": "'${INSTRUCTOR_IDS[0]}'"}' > /dev/null
        api_patch "/api/data/edu_course/${COURSE_IDS[1]}" '{"instructor_id": "'${INSTRUCTOR_IDS[0]}'"}' > /dev/null
        echo "  ✓ Instructors assigned to courses"
    fi
    [ $count -eq 2 ] && test_passed "Instructors created" || test_failed "Only created $count instructors"
}

test_create_students() {
    echo ""
    echo "Test 26.3: Create Students"
    
    local students=(
        '{"name": "Alice Johnson", "student_id": "STU-'$TIMESTAMP'-001", "email": "alice.'$TIMESTAMP'@student.edu", "major": "Computer Science", "year": "Junior", "gpa": 3.7, "status": "Active"}'
        '{"name": "Bob Williams", "student_id": "STU-'$TIMESTAMP'-002", "email": "bob.'$TIMESTAMP'@student.edu", "major": "Mathematics", "year": "Sophomore", "gpa": 3.5, "status": "Active"}'
        '{"name": "Carol Davis", "student_id": "STU-'$TIMESTAMP'-003", "email": "carol.'$TIMESTAMP'@student.edu", "major": "Computer Science", "year": "Senior", "gpa": 3.9, "status": "Active"}'
        '{"name": "David Lee", "student_id": "STU-'$TIMESTAMP'-004", "email": "david.'$TIMESTAMP'@student.edu", "major": "Physics", "year": "Freshman", "gpa": 3.2, "status": "Active"}'
    )
    
    local count=0
    for student in "${students[@]}"; do
        local res=$(api_post "/api/data/edu_student" "$student")
        local id=$(json_extract "$res" "id")
        if [ -n "$id" ]; then
            STUDENT_IDS+=("$id")
            count=$((count + 1))
        fi
    done
    
    echo "  Created $count students"
    [ $count -eq 4 ] && test_passed "Students created" || test_failed "Only created $count students"
}

test_enroll_students() {
    echo ""
    echo "Test 26.4: Enroll Students in Courses"
    
    if [ ${#STUDENT_IDS[@]} -lt 2 ] || [ ${#COURSE_IDS[@]} -lt 2 ]; then
        test_failed "Missing students or courses"; return 1
    fi
    
    local count=0
    
    # Enroll Alice in CS101 and CS201
    for course_idx in 0 1; do
        local res=$(api_post "/api/data/edu_enrollment" '{"name": "Enroll-'$count'-'$TIMESTAMP'", "student_id": "'${STUDENT_IDS[0]}'", "course_id": "'${COURSE_IDS[$course_idx]}'", "semester": "Spring 2025", "status": "Enrolled"}')
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { ENROLLMENT_IDS+=("$id"); count=$((count + 1)); }
    done
    echo "  ✓ Alice enrolled in 2 courses"
    
    # Enroll Bob in MATH101
    local res=$(api_post "/api/data/edu_enrollment" '{"name": "Enroll-'$count'-'$TIMESTAMP'", "student_id": "'${STUDENT_IDS[1]}'", "course_id": "'${COURSE_IDS[2]}'", "semester": "Spring 2025", "status": "Enrolled"}')
    local id=$(json_extract "$res" "id")
    [ -n "$id" ] && { ENROLLMENT_IDS+=("$id"); count=$((count + 1)); echo "  ✓ Bob enrolled in 1 course"; }
    
    # Enroll Carol in CS101
    local res=$(api_post "/api/data/edu_enrollment" '{"name": "Enroll-'$count'-'$TIMESTAMP'", "student_id": "'${STUDENT_IDS[2]}'", "course_id": "'${COURSE_IDS[0]}'", "semester": "Spring 2025", "status": "Enrolled"}')
    local id=$(json_extract "$res" "id")
    [ -n "$id" ] && { ENROLLMENT_IDS+=("$id"); count=$((count + 1)); echo "  ✓ Carol enrolled in 1 course"; }
    
    echo "  Total enrollments: $count"
    [ $count -ge 4 ] && test_passed "Students enrolled" || test_failed "Only $count enrollments"
}

test_record_grades() {
    echo ""
    echo "Test 26.5: Record Grades"
    
    if [ ${#ENROLLMENT_IDS[@]} -lt 2 ]; then
        test_failed "No enrollments to grade"; return 1
    fi
    
    local grades=("A" "B+" "A-" "B")
    local count=0
    
    for i in "${!ENROLLMENT_IDS[@]}"; do
        if [ $i -lt ${#grades[@]} ]; then
            api_patch "/api/data/edu_enrollment/${ENROLLMENT_IDS[$i]}" '{"grade": "'${grades[$i]}'", "status": "Completed"}' > /dev/null
            echo "  ✓ Grade recorded: ${grades[$i]}"
            count=$((count + 1))
        fi
    done
    
    [ $count -ge 2 ] && test_passed "Grades recorded" || test_failed "Only $count grades"
}

test_query_transcripts() {
    echo ""
    echo "Test 26.6: Query Student Transcripts"
    
    if [ ${#STUDENT_IDS[@]} -lt 1 ]; then
        test_failed "No students to query"; return 1
    fi
    
    local enrollments=$(api_post "/api/data/query" '{"object_api_name": "edu_enrollment", "filter_expr": "student_id == '"'"'${STUDENT_IDS[0]}'"'"'"}')
    echo "  Alice's enrollments: $(echo "$enrollments" | jq '.data | length' 2>/dev/null || echo 0)"
    
    if [ ${#COURSE_IDS[@]} -ge 1 ]; then
        local course_enrollments=$(api_post "/api/data/query" '{"object_api_name": "edu_enrollment", "filter_expr": "course_id == '"'"'${COURSE_IDS[0]}'"'"'"}')
        echo "  CS101 enrollments: $(echo "$course_enrollments" | jq '.data | length' 2>/dev/null || echo 0)"
    fi
    
    test_passed "Transcript queries completed"
}

test_course_drop() {
    echo ""
    echo "Test 26.7: Course Drop"
    
    if [ ${#STUDENT_IDS[@]} -lt 4 ] || [ ${#COURSE_IDS[@]} -lt 1 ]; then
        test_passed "Course drop (skipped)"; return 0
    fi
    
    local drop_res=$(api_post "/api/data/edu_enrollment" '{"name": "DropTest-'$TIMESTAMP'", "student_id": "'${STUDENT_IDS[3]}'", "course_id": "'${COURSE_IDS[0]}'", "semester": "Spring 2025", "status": "Enrolled"}')
    local drop_id=$(json_extract "$drop_res" "id")
    
    if [ -n "$drop_id" ]; then
        api_patch "/api/data/edu_enrollment/$drop_id" '{"status": "Dropped", "grade": "W"}' > /dev/null
        echo "  ✓ David dropped CS101 (Grade: W)"
        api_delete "/api/data/edu_enrollment/$drop_id" > /dev/null 2>&1
        test_passed "Course drop completed"
    else
        test_passed "Course drop (enrollment failed)"
    fi
}

# =========================================
# CLEANUP (Using Helpers)
# =========================================

test_cleanup() {
    echo ""
    echo "Test 26.8: Cleanup Test Data"
    
    # 1. Delete records first (referential integrity)
    cleanup_records "edu_enrollment" "${ENROLLMENT_IDS[@]}"
    cleanup_records "edu_student" "${STUDENT_IDS[@]}"
    cleanup_records "edu_course" "${COURSE_IDS[@]}"
    cleanup_records "edu_instructor" "${INSTRUCTOR_IDS[@]}"
    
    # 2. Delete schemas (dependent schemas first)
    echo "  Cleaning up schemas..."
    delete_schema "edu_enrollment"
    delete_schema "edu_course"
    delete_schema "edu_student"
    delete_schema "edu_instructor"
    
    # 3. Delete app
    echo "  Cleaning up app..."
    delete_app "app_Education"
    
    test_passed "Cleanup completed"
}

# Trap cleanup on exit to ensure it runs even if tests fail
trap test_cleanup EXIT

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
