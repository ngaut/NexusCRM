#!/bin/bash
set -e

# Suite 46: Referential Integrity
# Tests relationship constraints (Restrict, Cascade)

SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"

test_cleanup() {
    echo "🧹 Cleaning up Referential Integrity test data..."
    api_login "admin@test.com" "Admin123!"
    
    delete_schema "ri_child_cascade" 2>/dev/null || true
    delete_schema "ri_child_restrict" 2>/dev/null || true
    delete_schema "ri_parent" 2>/dev/null || true
}

trap test_cleanup EXIT

# Login as Admin
api_login "admin@test.com" "Admin123!"

echo "🔗 Setting up test schemas for Referential Integrity..."

# Create Parent Object
ensure_schema "ri_parent" "RI Parent"

# Create Child Object with Cascade Delete
ensure_schema "ri_child_cascade" "RI Child Cascade"
# Add Lookup with Cascade
# Note: delete_rule must be passed in JSON. "delete_rule": "Cascade"
add_field "ri_child_cascade" "parent_id" "Parent" "Lookup" "true" "{\"reference_to\": [\"ri_parent\"], \"delete_rule\": \"Cascade\"}"

# Create Child Object with Restrict Delete
ensure_schema "ri_child_restrict" "RI Child Restrict"
# Add Lookup with Restrict
add_field "ri_child_restrict" "parent_id" "Parent" "Lookup" "true" "{\"reference_to\": [\"ri_parent\"], \"delete_rule\": \"Restrict\"}"

sleep 2 # Wait for schema cache

# --- TEST 1: CASCADE DELETE ---
echo "🧪 Test 1: Verify Cascade Delete..."

# Create Parent
PARENT1_RES=$(api_post "/api/data/ri_parent" "{\"name\": \"Parent For Cascade\"}")
PARENT1_ID=$(json_extract "$PARENT1_RES" "id")
echo "   Parent created: $PARENT1_ID"

# Create Child linked to Parent
CHILD1_RES=$(api_post "/api/data/ri_child_cascade" "{\"name\": \"Child Cascade 1\", \"parent_id\": \"$PARENT1_ID\"}")
CHILD1_ID=$(json_extract "$CHILD1_RES" "id")
echo "   Child created: $CHILD1_ID"

# Delete Parent
echo "   Deleting parent..."
api_delete "/api/data/ri_parent/$PARENT1_ID"

# Verify Child is deleted
echo "   Verifying child deletion..."
CHILD_CHECK=$(api_get "/api/data/ri_child_cascade/$CHILD1_ID")

if echo "$CHILD_CHECK" | jq -e '.id' >/dev/null; then
    echo "❌ Child should have been deleted but exists: $CHILD_CHECK"
    exit 1
fi
# Need to distinguish between 404 (success) and other errors.
# api_get returns the body. If row not found, usually 404 error json.

if echo "$CHILD_CHECK" | grep -q "not found" || echo "$CHILD_CHECK" | grep -q "\"error\""; then
    echo "✅ Child correctly deleted (Cascade worked)"
else
    # Double check by querying
    COUNT=$(api_post "/api/data/query" "{\"object_api_name\": \"ri_child_cascade\", \"filters\": [{\"field\": \"id\", \"operator\": \"=\", \"value\": \"$CHILD1_ID\"}]}" | jq '.records | length')
    if [[ "$COUNT" -eq 0 ]]; then
         echo "✅ Child correctly deleted (Cascade confirmed via query)"
    else
         echo "❌ Child still exists after parent delete!"
         exit 1
    fi
fi

# --- TEST 2: RESTRICT DELETE ---
echo "🧪 Test 2: Verify Restrict Delete..."

# Create Parent
PARENT2_RES=$(api_post "/api/data/ri_parent" "{\"name\": \"Parent For Restrict\"}")
PARENT2_ID=$(json_extract "$PARENT2_RES" "id")
echo "   Parent created: $PARENT2_ID"

# Create Child linked to Parent
CHILD2_RES=$(api_post "/api/data/ri_child_restrict" "{\"name\": \"Child Restrict 1\", \"parent_id\": \"$PARENT2_ID\"}")
CHILD2_ID=$(json_extract "$CHILD2_RES" "id")
echo "   Child created: $CHILD2_ID"

# Try to Delete Parent (Should Fail)
echo "   Attempting to delete parent (should fail)..."
DELETE_RES=$(api_delete "/api/data/ri_parent/$PARENT2_ID")

# Verify Failure
if echo "$DELETE_RES" | grep -q "error"; then
     # Check for specific restrict message if possible (usually "validation error" or "foreign key constraint")
     echo "✅ Delete blocked: $DELETE_RES"
else
     echo "❌ Parent was deleted but should have been restricted! Response: $DELETE_RES"
     exit 1
fi

# Clean up child first then parent
echo "   Cleaning up manually..."
api_delete "/api/data/ri_child_restrict/$CHILD2_ID"
api_delete "/api/data/ri_parent/$PARENT2_ID"

echo "✅ Suite 46 Passed! (Referential Integrity)"
