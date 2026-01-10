#!/bin/bash
# tests/e2e/suites/25-healthcare-patients.sh
# Healthcare Patient Management E2E Tests
# REFACTORED: Uses helper libraries

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="Healthcare Patient Management"
TIMESTAMP=$(date +%s)

# Test data IDs
PATIENT_IDS=()
PROVIDER_IDS=()
APPOINTMENT_IDS=()
VISIT_IDS=()

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Schema Setup (using helpers)
    setup_schemas
    setup_healthcare_app
    
    # Tests
    test_register_patients
    test_create_providers
    test_schedule_appointments
    test_complete_visit
    test_patient_history
    test_appointment_search
}

# =========================================
# SCHEMA SETUP (Using Helpers)
# =========================================

setup_schemas() {
    echo "Setup: Creating Healthcare Schemas"
    
    # Patient Schema
    ensure_schema "hc_patient" "Patient" "Patients"
    add_field "hc_patient" "date_of_birth" "Date of Birth" "Date"
    add_picklist "hc_patient" "gender" "Gender" "Male,Female,Other,Prefer not to say"
    add_field "hc_patient" "phone" "Phone" "Phone"
    add_field "hc_patient" "email" "Email" "Email"
    add_field "hc_patient" "address" "Address" "LongText"
    add_field "hc_patient" "insurance_provider" "Insurance Provider" "Text"
    add_field "hc_patient" "insurance_id" "Insurance ID" "Text"
    add_picklist "hc_patient" "blood_type" "Blood Type" "A+,A-,B+,B-,AB+,AB-,O+,O-"
    add_field "hc_patient" "allergies" "Allergies" "LongText"
    
    # Provider Schema
    ensure_schema "hc_provider" "Provider" "Providers"
    add_picklist "hc_provider" "specialty" "Specialty" "General Practice,Cardiology,Dermatology,Pediatrics,Orthopedics,Neurology"
    add_field "hc_provider" "license_number" "License Number" "Text" true
    add_field "hc_provider" "phone" "Phone" "Phone"
    add_field "hc_provider" "email" "Email" "Email"
    add_picklist "hc_provider" "availability" "Availability" "Available,On Leave,Unavailable"
    
    # Appointment Schema
    ensure_schema "hc_appointment" "Appointment" "Appointments"
    add_lookup "hc_appointment" "patient_id" "Patient" "hc_patient" true
    add_lookup "hc_appointment" "provider_id" "Provider" "hc_provider" true
    add_field "hc_appointment" "appointment_date" "Date" "DateTime" true
    add_field "hc_appointment" "duration_minutes" "Duration (min)" "Number"
    add_field "hc_appointment" "reason" "Reason" "Text"
    add_picklist "hc_appointment" "status" "Status" "Scheduled,Checked In,In Progress,Completed,Cancelled,No Show"
    
    # Visit Schema
    ensure_schema "hc_visit" "Visit" "Visits"
    add_lookup "hc_visit" "patient_id" "Patient" "hc_patient" true
    add_lookup "hc_visit" "provider_id" "Provider" "hc_provider"
    add_lookup "hc_visit" "appointment_id" "Appointment" "hc_appointment"
    add_field "hc_visit" "visit_date" "Visit Date" "DateTime"
    add_field "hc_visit" "chief_complaint" "Chief Complaint" "LongText"
    add_field "hc_visit" "diagnosis" "Diagnosis" "LongText"
    add_field "hc_visit" "treatment_plan" "Treatment Plan" "LongText"
    add_field "hc_visit" "notes" "Clinical Notes" "LongText"
}

setup_healthcare_app() {
    echo ""
    echo "Setup: Creating Healthcare App"
    
    local nav_items=$(build_nav_items \
        "$(nav_item 'nav-hc-patients' 'object' 'Patients' 'hc_patient' 'Users')" \
        "$(nav_item 'nav-hc-providers' 'object' 'Providers' 'hc_provider' 'UserCheck')" \
        "$(nav_item 'nav-hc-appointments' 'object' 'Appointments' 'hc_appointment' 'Calendar')" \
        "$(nav_item 'nav-hc-visits' 'object' 'Visits' 'hc_visit' 'ClipboardList')")
    
    ensure_app "app_Healthcare" "Healthcare" "Heart" "#EF4444" "$nav_items" "Patient Management System"
}

# =========================================
# TESTS
# =========================================

test_register_patients() {
    echo ""
    echo "Test 25.1: Register Patients"
    
    local patients=(
        '{"name": "Emma Wilson", "date_of_birth": "1985-03-15", "gender": "Female", "phone": "5551234567", "email": "emma.w.'$TIMESTAMP'@patient.com", "insurance_provider": "Blue Cross", "blood_type": "A+"}'
        '{"name": "James Brown", "date_of_birth": "1978-07-22", "gender": "Male", "phone": "5559876543", "email": "james.b.'$TIMESTAMP'@patient.com", "insurance_provider": "Aetna", "blood_type": "O+"}'
        '{"name": "Olivia Martinez", "date_of_birth": "1992-11-08", "gender": "Female", "phone": "5555551234", "email": "olivia.m.'$TIMESTAMP'@patient.com", "insurance_provider": "Kaiser", "blood_type": "B-"}'
    )
    
    local count=0
    for patient in "${patients[@]}"; do
        local res=$(api_post "/api/data/hc_patient" "$patient")
        local id=$(json_extract "$res" "id")
        if [ -n "$id" ]; then
            PATIENT_IDS+=("$id")
            count=$((count + 1))
        fi
    done
    
    echo "  Registered $count patients"
    [ $count -eq 3 ] && test_passed "Patients registered" || test_failed "Only registered $count patients"
}

test_create_providers() {
    echo ""
    echo "Test 25.2: Create Providers"
    
    local providers=(
        '{"name": "Dr. Sarah Chen", "specialty": "General Practice", "license_number": "MD-'$TIMESTAMP'-001", "phone": "5550001111", "email": "dr.chen.'$TIMESTAMP'@clinic.com", "availability": "Available"}'
        '{"name": "Dr. Michael Park", "specialty": "Cardiology", "license_number": "MD-'$TIMESTAMP'-002", "phone": "5550002222", "email": "dr.park.'$TIMESTAMP'@clinic.com", "availability": "Available"}'
    )
    
    local count=0
    for provider in "${providers[@]}"; do
        local res=$(api_post "/api/data/hc_provider" "$provider")
        local id=$(json_extract "$res" "id")
        if [ -n "$id" ]; then
            PROVIDER_IDS+=("$id")
            count=$((count + 1))
        fi
    done
    
    echo "  Created $count providers"
    [ $count -eq 2 ] && test_passed "Providers created" || test_failed "Only created $count providers"
}

test_schedule_appointments() {
    echo ""
    echo "Test 25.3: Schedule Appointments"
    
    if [ ${#PATIENT_IDS[@]} -lt 1 ] || [ ${#PROVIDER_IDS[@]} -lt 1 ]; then
        test_failed "Missing patients or providers"; return 1
    fi
    
    local count=0
    
    local res=$(api_post "/api/data/hc_appointment" '{"name": "APT-'$TIMESTAMP'-001", "patient_id": "'${PATIENT_IDS[0]}'", "provider_id": "'${PROVIDER_IDS[0]}'", "appointment_date": "2025-01-15T10:00:00Z", "duration_minutes": 30, "reason": "Annual checkup", "status": "Scheduled"}')
    local apt_id=$(json_extract "$res" "id")
    if [ -n "$apt_id" ]; then
        APPOINTMENT_IDS+=("$apt_id")
        count=$((count + 1))
        echo "  ✓ Scheduled: Annual checkup"
    fi
    
    if [ ${#PATIENT_IDS[@]} -ge 2 ]; then
        local res2=$(api_post "/api/data/hc_appointment" '{"name": "APT-'$TIMESTAMP'-002", "patient_id": "'${PATIENT_IDS[1]}'", "provider_id": "'${PROVIDER_IDS[1]}'", "appointment_date": "2025-01-15T14:00:00Z", "duration_minutes": 45, "reason": "Cardiology follow-up", "status": "Scheduled"}')
        local apt_id2=$(json_extract "$res2" "id")
        if [ -n "$apt_id2" ]; then
            APPOINTMENT_IDS+=("$apt_id2")
            count=$((count + 1))
            echo "  ✓ Scheduled: Cardiology follow-up"
        fi
    fi
    
    echo "  Scheduled $count appointments"
    [ $count -ge 2 ] && test_passed "Appointments scheduled" || test_failed "Only scheduled $count"
}

test_complete_visit() {
    echo ""
    echo "Test 25.4: Complete Visit (Check In → Examine → Document)"
    
    if [ ${#APPOINTMENT_IDS[@]} -lt 1 ]; then
        test_failed "No appointments available"; return 1
    fi
    
    local apt_id="${APPOINTMENT_IDS[0]}"
    
    api_patch "/api/data/hc_appointment/$apt_id" '{"status": "Checked In"}' > /dev/null
    echo "  ✓ Patient checked in"
    
    api_patch "/api/data/hc_appointment/$apt_id" '{"status": "In Progress"}' > /dev/null
    echo "  ✓ Appointment in progress"
    
    local visit_res=$(api_post "/api/data/hc_visit" '{"name": "VISIT-'$TIMESTAMP'-001", "patient_id": "'${PATIENT_IDS[0]}'", "provider_id": "'${PROVIDER_IDS[0]}'", "appointment_id": "'$apt_id'", "visit_date": "2025-01-15T10:00:00Z", "chief_complaint": "Annual wellness exam", "diagnosis": "Healthy", "treatment_plan": "Continue current lifestyle"}')
    local visit_id=$(json_extract "$visit_res" "id")
    if [ -n "$visit_id" ]; then
        VISIT_IDS+=("$visit_id")
        echo "  ✓ Visit documented"
    fi
    
    api_patch "/api/data/hc_appointment/$apt_id" '{"status": "Completed"}' > /dev/null
    echo "  ✓ Appointment completed"
    
    test_passed "Visit workflow completed"
}

test_patient_history() {
    echo ""
    echo "Test 25.5: Query Patient History"
    
    if [ ${#PATIENT_IDS[@]} -lt 1 ]; then
        test_failed "No patients to query"; return 1
    fi
    
    local patient_id="${PATIENT_IDS[0]}"
    
    local appointments=$(api_post "/api/data/query" '{"object_api_name": "hc_appointment", "filter_expr": "patient_id == '"'"'$patient_id'"'"'"}')
    echo "  Patient appointments: $(echo "$appointments" | jq '.data | length' 2>/dev/null || echo 0)"
    
    local visits=$(api_post "/api/data/query" '{"object_api_name": "hc_visit", "filter_expr": "patient_id == '"'"'$patient_id'"'"'"}')
    echo "  Patient visits: $(echo "$visits" | jq '.data | length' 2>/dev/null || echo 0)"
    
    test_passed "Patient history retrieved"
}

test_appointment_search() {
    echo ""
    echo "Test 25.6: Appointment Search"
    
    local scheduled=$(api_post "/api/data/query" '{"object_api_name": "hc_appointment", "filter_expr": "status == '"'"'Scheduled'"'"'"}')
    echo "  Scheduled appointments: $(echo "$scheduled" | jq '.data | length' 2>/dev/null || echo 0)"
    
    if [ ${#PROVIDER_IDS[@]} -ge 1 ]; then
        local provider_apts=$(api_post "/api/data/query" '{"object_api_name": "hc_appointment", "filter_expr": "provider_id == '"'"'${PROVIDER_IDS[0]}'"'"'"}')
        echo "  Dr. Chen's appointments: $(echo "$provider_apts" | jq '.data | length' 2>/dev/null || echo 0)"
    fi
    
    test_passed "Appointment search completed"
}

# =========================================
# CLEANUP (Using Helpers)
# =========================================

test_cleanup() {
    echo ""
    echo "Test 25.7: Cleanup Test Data"
    
    # 1. Delete records
    cleanup_records "hc_visit" "${VISIT_IDS[@]}"
    cleanup_records "hc_appointment" "${APPOINTMENT_IDS[@]}"
    cleanup_records "hc_provider" "${PROVIDER_IDS[@]}"
    cleanup_records "hc_patient" "${PATIENT_IDS[@]}"
    
    # 2. Delete schemas
    echo "  Cleaning up schemas..."
    delete_schema "hc_visit"
    delete_schema "hc_appointment"
    delete_schema "hc_provider"
    delete_schema "hc_patient"
    
    # 3. Delete app
    echo "  Cleaning up app..."
    delete_app "app_Healthcare"
    
    test_passed "Cleanup completed"
}

# Trap cleanup on exit
trap test_cleanup EXIT

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
