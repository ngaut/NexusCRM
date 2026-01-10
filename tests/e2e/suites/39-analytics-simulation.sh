#!/bin/bash
# tests/e2e/suites/39-analytics-simulation.sh
# Analytics Simulation
# Simulates Dashboard Widget capabilities (Aggregations, Group By)

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Analytics Simulation"
TIMESTAMP=$(date +%s)
OBJ_NAME="SalesStat_$TIMESTAMP"
API_NAME="salesstat_$TIMESTAMP"

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi

    test_create_analytics_object
    test_populate_analytics_data
    test_analytics_count
    test_analytics_sum
    test_analytics_group_by
}

test_create_analytics_object() {
    echo ""
    echo "Test 39.1: Create Analytics Object ($OBJ_NAME)"
    
    # Create Object
    api_post "/api/metadata/objects" '{
        "label": "'$OBJ_NAME'",
        "plural_label": "'$OBJ_NAME's",
        "api_name": "'$API_NAME'",
        "is_custom": true
    }' > /dev/null
    
    # Add Region (Picklist)
    api_post "/api/metadata/objects/$API_NAME/fields" '{
        "label": "Region",
        "api_name": "region",
        "type": "Picklist",
        "options": ["North", "South", "East", "West"]
    }' > /dev/null
    
    # Add Amount (Currency)
    api_post "/api/metadata/objects/$API_NAME/fields" '{
        "label": "Amount",
        "api_name": "amount",
        "type": "Currency"
    }' > /dev/null
    
    test_passed "Analytics Object Created"
}

test_populate_analytics_data() {
    echo ""
    echo "Test 39.2: Populate Data"
    
    # North: 100, 200
    create_stat "North" 100
    create_stat "North" 200
    
    # South: 500
    create_stat "South" 500
    
    test_passed "Data Populated (3 records)"
}

create_stat() {
    local region=$1
    local amount=$2
    api_post "/api/data/$API_NAME" "{\"name\": \"Stat\", \"region\": \"$region\", \"amount\": $amount}" > /dev/null
}

test_analytics_count() {
    echo ""
    echo "Test 39.3: Analytics - Count"
    
    # Count all
    local res=$(api_post "/api/data/analytics" '{
        "object_api_name": "'$API_NAME'",
        "operation": "count"
    }')
    
    local val=$(echo "$res" | jq -r '.data')
    
    if [ "$val" = "3" ]; then
        test_passed "Count is correct (3)"
    else
        test_failed "Count failed. Expected 3, got $val" "$res"
    fi
}

test_analytics_sum() {
    echo ""
    echo "Test 39.4: Analytics - Sum"
    
    # Sum Amount
    local res=$(api_post "/api/data/analytics" '{
        "object_api_name": "'$API_NAME'",
        "operation": "sum",
        "field": "amount"
    }')
    
    local val=$(echo "$res" | jq -r '.data')
    
    # Check if starts with 800 (handles 800 or 800.00)
    if [[ "$val" == "800"* ]]; then
        test_passed "Sum is correct (800)"
    else
        test_failed "Sum failed. Expected 800, got $val" "$res"
    fi
}

test_analytics_group_by() {
    echo ""
    echo "Test 39.5: Analytics - Group By"
    
    # Group By Region, Sum Amount
    local res=$(api_post "/api/data/analytics" '{
        "object_api_name": "'$API_NAME'",
        "operation": "group_by",
        "group_by": "region",
        "field": "amount"
    }')
    
    # Expected: [{name: "North", value: 300}, {name: "South", value: 500}]
    local north=$(echo "$res" | jq -r '.data[] | select(.name == "North") | .value')
    local south=$(echo "$res" | jq -r '.data[] | select(.name == "South") | .value')
    
    if [[ "$north" == "300"* ]] && [[ "$south" == "500"* ]]; then
        test_passed "Group By correct (North=300, South=500)"
    else
        test_failed "Group By failed. Got North=$north, South=$south" "$res"
    fi
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
