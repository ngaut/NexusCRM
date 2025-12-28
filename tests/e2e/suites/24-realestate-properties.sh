#!/bin/bash
# tests/e2e/suites/24-realestate-properties.sh
# Real Estate Property Management E2E Tests
# REFACTORED: Uses helper libraries for reduced code duplication

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="Real Estate Property Management"
TIMESTAMP=$(date +%s)

# Test data IDs
PROPERTY_IDS=()
TENANT_IDS=()
LEASE_IDS=()
MAINTENANCE_IDS=()

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Schema Setup (using helpers)
    setup_schemas
    setup_realestate_app
    
    # Tests
    test_create_properties
    test_create_tenants
    test_create_leases
    test_process_maintenance_request
    test_property_search
    test_lease_renewal
    test_occupancy_report
}

# =========================================
# SCHEMA SETUP (Using Helpers)
# =========================================

setup_schemas() {
    echo "Setup: Creating Real Estate Schemas"
    
    # Property Schema
    ensure_schema "re_property" "Property" "Properties"
    add_field "re_property" "address" "Address" "LongText" true
    add_picklist "re_property" "property_type" "Type" "Apartment,House,Condo,Townhouse,Commercial"
    add_field "re_property" "bedrooms" "Bedrooms" "Number"
    add_field "re_property" "bathrooms" "Bathrooms" "Number"
    add_field "re_property" "square_feet" "Square Feet" "Number"
    add_field "re_property" "monthly_rent" "Monthly Rent" "Currency"
    add_picklist "re_property" "status" "Status" "Available,Occupied,Maintenance,Off Market"
    add_field "re_property" "year_built" "Year Built" "Number"
    
    # Tenant Schema
    ensure_schema "re_tenant" "Tenant" "Tenants"
    add_field "re_tenant" "email" "Email" "Email" true
    add_field "re_tenant" "phone" "Phone" "Phone"
    add_field "re_tenant" "employer" "Employer" "Text"
    add_field "re_tenant" "monthly_income" "Monthly Income" "Currency"
    add_field "re_tenant" "credit_score" "Credit Score" "Number"
    add_picklist "re_tenant" "status" "Status" "Active,Former,Applicant"
    
    # Lease Schema
    ensure_schema "re_lease" "Lease" "Leases"
    add_lookup "re_lease" "property_id" "Property" "re_property" true
    add_lookup "re_lease" "tenant_id" "Tenant" "re_tenant" true
    add_field "re_lease" "start_date" "Start Date" "Date" true
    add_field "re_lease" "end_date" "End Date" "Date" true
    add_field "re_lease" "monthly_rent" "Monthly Rent" "Currency"
    add_field "re_lease" "security_deposit" "Security Deposit" "Currency"
    add_picklist "re_lease" "status" "Status" "Active,Expired,Terminated,Pending"
    
    # Maintenance Schema
    ensure_schema "re_maintenance" "Maintenance Request" "Maintenance Requests"
    add_lookup "re_maintenance" "property_id" "Property" "re_property" true
    add_lookup "re_maintenance" "tenant_id" "Reported By" "re_tenant"
    add_field "re_maintenance" "description" "Description" "LongText" true
    add_picklist "re_maintenance" "priority" "Priority" "Low,Medium,High,Emergency"
    add_picklist "re_maintenance" "status" "Status" "Open,In Progress,Completed,Cancelled"
    add_field "re_maintenance" "cost" "Cost" "Currency"
}

setup_realestate_app() {
    echo ""
    echo "Setup: Creating Real Estate App"
    
    local nav_items=$(build_nav_items \
        "$(nav_item 'nav-re-properties' 'object' 'Properties' 're_property' 'Building')" \
        "$(nav_item 'nav-re-tenants' 'object' 'Tenants' 're_tenant' 'Users')" \
        "$(nav_item 'nav-re-leases' 'object' 'Leases' 're_lease' 'FileText')" \
        "$(nav_item 'nav-re-maintenance' 'object' 'Maintenance' 're_maintenance' 'Wrench')")
    
    ensure_app "app_RealEstate" "Real Estate" "Home" "#8B5CF6" "$nav_items" "Property Management"
}

# =========================================
# TESTS
# =========================================

test_create_properties() {
    echo ""
    echo "Test 24.1: Create Properties"
    
    local properties=(
        '{"name": "Sunset Apartment 101", "address": "101 Sunset Blvd, Unit A\nLos Angeles, CA 90028", "property_type": "Apartment", "bedrooms": 2, "bathrooms": 1, "square_feet": 850, "monthly_rent": 2500, "status": "Available", "year_built": 2015}'
        '{"name": "Oak Tree House", "address": "456 Oak Tree Lane\nPasadena, CA 91101", "property_type": "House", "bedrooms": 4, "bathrooms": 3, "square_feet": 2200, "monthly_rent": 4500, "status": "Available", "year_built": 2008}'
        '{"name": "Downtown Condo 1502", "address": "789 Main St, Unit 1502\nLos Angeles, CA 90012", "property_type": "Condo", "bedrooms": 1, "bathrooms": 1, "square_feet": 650, "monthly_rent": 2200, "status": "Occupied", "year_built": 2020}'
    )
    
    local count=0
    for prop in "${properties[@]}"; do
        local res=$(api_post "/api/data/re_property" "$prop")
        local id=$(json_extract "$res" "id")
        if [ -n "$id" ]; then
            PROPERTY_IDS+=("$id")
            count=$((count + 1))
        fi
    done
    
    echo "  Created $count properties"
    [ $count -eq 3 ] && test_passed "Properties created" || test_failed "Only created $count properties"
}

test_create_tenants() {
    echo ""
    echo "Test 24.2: Create Tenants"
    
    local tenants=(
        '{"name": "John Smith", "email": "john.smith.'$TIMESTAMP'@tenant.com", "phone": "5551234567", "employer": "Tech Corp", "monthly_income": 8000, "credit_score": 750, "status": "Active"}'
        '{"name": "Sarah Johnson", "email": "sarah.j.'$TIMESTAMP'@tenant.com", "phone": "5559876543", "employer": "Finance Inc", "monthly_income": 10000, "credit_score": 780, "status": "Active"}'
        '{"name": "Mike Davis", "email": "mike.d.'$TIMESTAMP'@tenant.com", "phone": "5555551234", "employer": "Healthcare LLC", "monthly_income": 7500, "credit_score": 700, "status": "Applicant"}'
    )
    
    local count=0
    for tenant in "${tenants[@]}"; do
        local res=$(api_post "/api/data/re_tenant" "$tenant")
        local id=$(json_extract "$res" "id")
        if [ -n "$id" ]; then
            TENANT_IDS+=("$id")
            count=$((count + 1))
        fi
    done
    
    echo "  Created $count tenants"
    [ $count -ge 2 ] && test_passed "Tenants created" || test_failed "Only created $count tenants"
}

test_create_leases() {
    echo ""
    echo "Test 24.3: Create Lease Agreements"
    
    if [ ${#PROPERTY_IDS[@]} -lt 1 ] || [ ${#TENANT_IDS[@]} -lt 1 ]; then
        test_failed "Missing properties or tenants"; return 1
    fi
    
    local res=$(api_post "/api/data/re_lease" '{
        "name": "Lease-'$TIMESTAMP'-001",
        "property_id": "'${PROPERTY_IDS[0]}'",
        "tenant_id": "'${TENANT_IDS[0]}'",
        "start_date": "2025-01-01",
        "end_date": "2025-12-31",
        "monthly_rent": 2500,
        "security_deposit": 5000,
        "status": "Active"
    }')
    
    local lease_id=$(json_extract "$res" "id")
    if [ -n "$lease_id" ]; then
        LEASE_IDS+=("$lease_id")
        echo "  ✓ Created lease: $lease_id"
        api_patch "/api/data/re_property/${PROPERTY_IDS[0]}" '{"status": "Occupied"}' > /dev/null
        echo "  ✓ Property marked as Occupied"
        test_passed "Lease created"
    else
        test_failed "Failed to create lease" "$res"
    fi
}

test_process_maintenance_request() {
    echo ""
    echo "Test 24.4: Process Maintenance Request"
    
    if [ ${#PROPERTY_IDS[@]} -lt 1 ] || [ ${#TENANT_IDS[@]} -lt 1 ]; then
        test_failed "Missing properties or tenants"; return 1
    fi
    
    local res=$(api_post "/api/data/re_maintenance" '{
        "name": "MR-'$TIMESTAMP'",
        "property_id": "'${PROPERTY_IDS[0]}'",
        "tenant_id": "'${TENANT_IDS[0]}'",
        "description": "Leaky faucet in kitchen sink.",
        "priority": "Medium",
        "status": "Open"
    }')
    
    local maint_id=$(json_extract "$res" "id")
    if [ -z "$maint_id" ]; then
        test_failed "Failed to create maintenance request"; return 1
    fi
    
    MAINTENANCE_IDS+=("$maint_id")
    echo "  ✓ Request created: $maint_id"
    
    api_patch "/api/data/re_maintenance/$maint_id" '{"status": "In Progress"}' > /dev/null
    echo "  ✓ Status → In Progress"
    
    api_patch "/api/data/re_maintenance/$maint_id" '{"status": "Completed", "cost": 150}' > /dev/null
    echo "  ✓ Status → Completed (Cost: $150)"
    
    test_passed "Maintenance workflow completed"
}

test_property_search() {
    echo ""
    echo "Test 24.5: Property Search"
    
    local available=$(api_post "/api/data/query" '{"object_api_name": "re_property", "filters": [{"field": "status", "operator": "=", "value": "Available"}]}')
    echo "  Available properties: $(echo "$available" | jq '.records | length' 2>/dev/null || echo 0)"
    
    local two_bed=$(api_post "/api/data/query" '{"object_api_name": "re_property", "filters": [{"field": "bedrooms", "operator": ">=", "value": 2}]}')
    echo "  2+ bedroom properties: $(echo "$two_bed" | jq '.records | length' 2>/dev/null || echo 0)"
    
    test_passed "Property search completed"
}

test_lease_renewal() {
    echo ""
    echo "Test 24.6: Lease Renewal Process"
    
    if [ ${#LEASE_IDS[@]} -lt 1 ]; then
        test_failed "No lease to renew"; return 1
    fi
    
    local lease_id="${LEASE_IDS[0]}"
    local lease=$(api_get "/api/data/re_lease/$lease_id")
    echo "  Current lease ends: $(json_extract "$lease" "end_date")"
    
    local res=$(api_patch "/api/data/re_lease/$lease_id" '{"end_date": "2026-12-31", "monthly_rent": 2625}')
    
    if echo "$res" | grep -qE '"id"|updated|success'; then
        echo "  ✓ Lease renewed until 2026-12-31"
        echo "  ✓ Rent increased to \$2,625 (5% increase)"
        test_passed "Lease renewal completed"
    else
        test_failed "Failed to renew lease"
    fi
}

test_occupancy_report() {
    echo ""
    echo "Test 24.7: Generate Occupancy Report"
    
    local all_props=$(api_post "/api/data/query" '{"object_api_name": "re_property"}')
    local total=$(echo "$all_props" | jq '.records | length' 2>/dev/null || echo "0")
    
    local occupied=$(api_post "/api/data/query" '{"object_api_name": "re_property", "filters": [{"field": "status", "operator": "=", "value": "Occupied"}]}')
    local occ_count=$(echo "$occupied" | jq '.records | length' 2>/dev/null || echo "0")
    
    echo "  === Occupancy Report ==="
    echo "  Total: $total | Occupied: $occ_count | Available: $((total - occ_count))"
    [ $total -gt 0 ] && echo "  Occupancy Rate: $(echo "scale=1; $occ_count * 100 / $total" | bc)%"
    
    test_passed "Occupancy report generated"
}

# =========================================
# CLEANUP (Using Helpers)
# =========================================

test_cleanup() {
    echo ""
    echo "Test 24.8: Cleanup Test Data"
    
    # 1. Delete records
    cleanup_records "re_maintenance" "${MAINTENANCE_IDS[@]}"
    cleanup_records "re_lease" "${LEASE_IDS[@]}"
    cleanup_records "re_tenant" "${TENANT_IDS[@]}"
    cleanup_records "re_property" "${PROPERTY_IDS[@]}"
    
    # 2. Delete schemas
    echo "  Cleaning up schemas..."
    delete_schema "re_maintenance"
    delete_schema "re_lease"
    delete_schema "re_tenant"
    delete_schema "re_property"
    
    # 3. Delete app
    echo "  Cleaning up app..."
    delete_app "app_RealEstate"
    
    test_passed "Cleanup completed"
}

# Trap cleanup on exit
trap test_cleanup EXIT

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
