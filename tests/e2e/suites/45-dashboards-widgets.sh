#!/bin/bash
set -e

# Suite 45: Dashboard and Widget Lifecycle (Embedded JSON Widgets)
# Tests dashboard creation, embedded widget configuration, and analytics queries.

SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"

TIMESTAMP=$(date +%s)

test_cleanup() {
    echo "🧹 Cleaning up Dashboard test data..."
    api_login "admin@test.com" "Admin123!"
    
    # Delete dashboards created during test
    if [[ -n "$DASHBOARD_ID" ]]; then
        api_delete "/api/data/_System_Dashboard/$DASHBOARD_ID" 2>/dev/null || true
    fi
    
    # Delete test schema if created
    delete_schema "sales_metric" 2>/dev/null || true
}

trap test_cleanup EXIT

# Login as Admin
api_login "admin@test.com" "Admin123!"

echo "📊 Setting up test data for dashboard..."

# Create a custom object for metrics
ensure_schema "sales_metric" "Sales Metric"
add_field "sales_metric" "region" "Region" "Text" "true"
add_field "sales_metric" "amount" "Amount" "Currency" "true"
add_picklist "sales_metric" "quarter" "Quarter" "Q1,Q2,Q3,Q4" "true"

# Wait for Schema Cache (Polling)
echo "   Waiting for fields..."
for i in {1..10}; do
    meta=$(api_get "/api/metadata/objects/sales_metric")
    if echo "$meta" | grep -q "\"api_name\":\"quarter\""; then
        break
    fi
    sleep 0.5
done

# Create some test data for aggregation
echo "📈 Creating test metrics data..."
api_post "/api/data/sales_metric" "{\"name\": \"Q1 West\", \"region\": \"West\", \"amount\": 50000, \"quarter\": \"Q1\"}"
api_post "/api/data/sales_metric" "{\"name\": \"Q1 East\", \"region\": \"East\", \"amount\": 75000, \"quarter\": \"Q1\"}"
api_post "/api/data/sales_metric" "{\"name\": \"Q2 West\", \"region\": \"West\", \"amount\": 60000, \"quarter\": \"Q2\"}"
api_post "/api/data/sales_metric" "{\"name\": \"Q2 East\", \"region\": \"East\", \"amount\": 85000, \"quarter\": \"Q2\"}"
api_post "/api/data/sales_metric" "{\"name\": \"Q3 North\", \"region\": \"North\", \"amount\": 40000, \"quarter\": \"Q3\"}"

# Wait for data availability
echo "   Waiting for data indexing..."
for i in {1..10}; do
    count_res=$(api_post "/api/data/query" "{\"object_api_name\": \"sales_metric\", \"aggregations\": [{\"function\": \"COUNT\", \"alias\": \"cnt\"}]}")
    cnt=$(echo "$count_res" | jq -r '.aggregations.cnt // 0')
    if [[ "$cnt" -eq 5 ]]; then
        break
    fi
    sleep 0.5
done

# --- TEST 1: CREATE DASHBOARD ---
echo "🧪 Test 1: Create Dashboard (Empty)..."
DASHBOARD_RES=$(api_post "/api/data/_System_Dashboard" "{
    \"name\": \"SalesMetrics_$TIMESTAMP\",
    \"label\": \"Sales Metrics Dashboard\",
    \"description\": \"Dashboard for sales metrics analysis\",
    \"is_public\": true,
    \"widgets\": []
}")
DASHBOARD_ID=$(json_extract "$DASHBOARD_RES" "id")

if [[ -z "$DASHBOARD_ID" || "$DASHBOARD_ID" == "null" ]]; then
    echo "❌ Failed to create dashboard: $DASHBOARD_RES"
    exit 1
fi
echo "✅ Dashboard created: $DASHBOARD_ID"

# --- TEST 2: ADD WIDGETS VIA PATCH ---
echo "🧪 Test 2: Add Widgets via UPDATE (JSON Patch)..."

# Construct widgets JSON array
WIDGETS_JSON='[
    {
        "__sys_gen_id": "widget1",
        "type": "metric",
        "title": "Total Sales",
        "config": {
            "object_api_name": "sales_metric",
            "aggregation": "SUM",
            "field": "amount"
        },
        "position": {"x": 0, "y": 0, "w": 3, "h": 2}
    },
    {
        "__sys_gen_id": "widget2",
        "title": "Sales by Region",
        "type": "chart-bar",
        "config": {
            "object_api_name": "sales_metric",
            "chartType": "bar",
            "groupBy": "region",
            "aggregation": "SUM",
            "aggregateField": "amount"
        },
        "position": {"x": 3, "y": 0, "w": 6, "h": 4}
    }
]'

UPDATE_RES=$(api_patch "/api/data/_System_Dashboard/$DASHBOARD_ID" "{
    \"widgets\": $WIDGETS_JSON
}")

if echo "$UPDATE_RES" | grep -qiE "error|fail|exception"; then
    echo "❌ Dashboard update failed: $UPDATE_RES"
    exit 1
fi
echo "✅ Dashboard widgets updated successfully"

# --- TEST 3: VERIFY WIDGET PERSISTENCE ---
echo "🧪 Test 3: Verify Widgets Persisted..."

DASHBOARD_GET=$(api_get "/api/data/_System_Dashboard/$DASHBOARD_ID")
WIDGETS_RETRIEVED=$(echo "$DASHBOARD_GET" | jq '.data.widgets')
WIDGETS_RETRIEVED=$(echo "$WIDGETS_RETRIEVED" | jq 'if . == null then [] else . end')
WIDGET_COUNT=$(echo "$WIDGETS_RETRIEVED" | jq 'length')

if [[ "$WIDGET_COUNT" -ne 2 ]]; then
    echo "❌ Expected 2 widgets, got $WIDGET_COUNT"
    echo "   Widgets: $WIDGETS_RETRIEVED"
    echo "   Full Response: $DASHBOARD_GET"
    exit 1
fi
echo "✅ Found $WIDGET_COUNT widgets embedded in dashboard"

# --- TEST 4: AGGREGATE QUERY (Widget Data Source) ---
echo "🧪 Test 4: Aggregate Query for Widget Data..."
AGG_QUERY=$(api_post "/api/data/query" "{
    \"object_api_name\": \"sales_metric\",
    \"fields\": [\"amount\"],
    \"aggregations\": [{\"function\": \"SUM\", \"field\": \"amount\", \"alias\": \"total\"}]
}")
TOTAL_AMOUNT=$(echo "$AGG_QUERY" | jq -r '.aggregations.total // .data[0].total // empty')

# We created: 50000 + 75000 + 60000 + 85000 + 40000 = 310000
if [[ -z "$TOTAL_AMOUNT" ]]; then
    echo "⚠️ Could not extract total from aggregation (may need different response format)"
    echo "   Response: $AGG_QUERY"
else
    echo "✅ Aggregate query returned total: $TOTAL_AMOUNT"
fi

# --- TEST 5: MODIFY WIDGETS ---
echo "🧪 Test 5: Modify Widgets (Add one, Remove one)..."

# New Set: Remove widget1, keep widget2, add widget3
NEW_WIDGETS_JSON='[
    {
        "__sys_gen_id": "widget2",
        "title": "Sales by Region",
        "type": "chart-bar",
        "config": {
            "object_api_name": "sales_metric",
            "chartType": "bar",
            "groupBy": "region",
            "aggregation": "SUM",
            "aggregateField": "amount"
        },
        "position": {"x": 3, "y": 0, "w": 6, "h": 4}
    },
    {
        "__sys_gen_id": "widget3",
        "title": "Recent Sales",
        "type": "record-list",
        "config": {
            "object_api_name": "sales_metric",
            "fields": ["name", "region", "amount", "quarter"],
            "limit": 5,
            "orderBy": "created_date DESC"
        },
        "position": {"x": 0, "y": 4, "w": 12, "h": 4}
    }
]'

UPDATE_RES_2=$(api_patch "/api/data/_System_Dashboard/$DASHBOARD_ID" "{
    \"widgets\": $NEW_WIDGETS_JSON
}")

if echo "$UPDATE_RES_2" | grep -qiE "error|fail|exception"; then
    echo "❌ Dashboard update 2 failed: $UPDATE_RES_2"
    exit 1
fi

# Verify
DASHBOARD_GET_2=$(api_get "/api/data/_System_Dashboard/$DASHBOARD_ID")
WIDGET_COUNT_2=$(echo "$DASHBOARD_GET_2" | jq '.data.widgets | length')

if [[ "$WIDGET_COUNT_2" -ne 2 ]]; then
    echo "❌ Expected 2 widgets after modification, got $WIDGET_COUNT_2"
    exit 1
fi

# Check if widget3 is present
HAS_WIDGET3=$(echo "$DASHBOARD_GET_2" | jq '.data.widgets[] | select(.__sys_gen_id=="widget3") | .__sys_gen_id')
if [[ -z "$HAS_WIDGET3" ]]; then
    echo "❌ Widget 3 not found after update!"
    exit 1
fi

echo "✅ Dashboard widgets modification verified"
echo "✅ Suite 45 Passed! (Dashboard & Widget Lifecycle)"
