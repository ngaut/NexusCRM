#!/bin/bash
set -e

# Suite 44: Complex Security Enforcement
# Tests OWD Private, Role Hierarchy, and Sharing Rules.

SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"

TIMESTAMP=$(date +%s)

test_cleanup() {
    echo "🧹 Cleaning up Security test data..."
    api_login "admin@test.com" "Admin123!"
    delete_schema "salary"
    # Individual record cleanup of users/roles is hard without IDs, 
    # but wipe_db is used between suites in some environments.
    # For now, we mainly care about the custom object.
}

trap test_cleanup EXIT

# 1. Login as Admin and setup roles/users
api_login "admin@test.com" "Admin123!"
ADMIN_TOKEN=$TOKEN

echo "👥 Setting up Roles..."
# Create HR Manager Role
ROLE_MGR_RES=$(api_post "/api/data/_System_Role" "{\"name\": \"HR_Manager_$TIMESTAMP\", \"label\": \"HR Manager\", \"description\": \"HR Manager Role\"}")
ROLE_MGR_ID=$(json_extract "$ROLE_MGR_RES" "id")

# Create HR Assistant Role (Subordinate to HR Manager)
ROLE_AST_RES=$(api_post "/api/data/_System_Role" "{\"name\": \"HR_Assistant_$TIMESTAMP\", \"label\": \"HR Assistant\", \"description\": \"HR Assistant Role\", \"parent_role_id\": \"$ROLE_MGR_ID\"}")
ROLE_AST_ID=$(json_extract "$ROLE_AST_RES" "id")

# Trigger role hierarchy cache refresh by hitting permissions endpoint
echo "   Manager Role ID: $ROLE_MGR_ID"
echo "   Assistant Role ID: $ROLE_AST_ID"

# Poll for role hierarchy update
echo "   Polling for role hierarchy consistency..."
for i in {1..10}; do
    me=$(api_get "/api/auth/permissions/me")
    if echo "$me" | grep -q "\"role_id\":\""; then
        break
    fi
    sleep 0.5
done

echo "👤 Creating Users..."
# Alice (Manager)
ALICE_EMAIL="alice.$TIMESTAMP@test.com"
ALICE_RES=$(api_post "/api/data/_System_User" "{\"username\": \"$ALICE_EMAIL\", \"email\": \"$ALICE_EMAIL\", \"password\": \"Password123!\", \"profile_id\": \"standard_user\", \"role_id\": \"$ROLE_MGR_ID\", \"first_name\": \"Alice\", \"last_name\": \"Manager\"}")
ALICE_ID=$(json_extract "$ALICE_RES" "id")
if [[ -z "$ALICE_ID" || "$ALICE_ID" == "null" ]]; then
    echo "❌ Failed to create Alice: $ALICE_RES"
    exit 1
fi

# Bob (Assistant)
BOB_EMAIL="bob.$TIMESTAMP@test.com"
BOB_RES=$(api_post "/api/data/_System_User" "{\"username\": \"$BOB_EMAIL\", \"email\": \"$BOB_EMAIL\", \"password\": \"Password123!\", \"profile_id\": \"standard_user\", \"role_id\": \"$ROLE_AST_ID\", \"first_name\": \"Bob\", \"last_name\": \"Assistant\"}")
BOB_ID=$(json_extract "$BOB_RES" "id")
if [[ -z "$BOB_ID" || "$BOB_ID" == "null" ]]; then
    echo "❌ Failed to create Bob: $BOB_RES"
    exit 1
fi

echo "🔐 Creating 'Salary' object (OWD: Private)..."
# Use api_post directly because ensure_schema might not support sharing_model yet
SALARY_OBJ=$(api_post "/api/metadata/objects" "{\"api_name\": \"salary\", \"label\": \"Salary\", \"plural_label\": \"Salaries\", \"sharing_model\": \"Private\", \"is_custom\": true}")
add_field "salary" "amount" "Amount" "Currency" "true"

# Grant CRUD permissions to standard_user on salary object
api_post "/api/data/_System_ObjectPerms" "{\"profile_id\": \"standard_user\", \"object_api_name\": \"salary\", \"allow_read\": true, \"allow_create\": true, \"allow_edit\": true, \"allow_delete\": true, \"view_all\": false, \"modify_all\": false}"

# Wait for cache/schema propagation
echo "   Waiting for 'salary' object metadata..."
for i in {1..10}; do
    meta=$(api_get "/api/metadata/objects/salary")
    if echo "$meta" | grep -q "\"api_name\":\"salary\""; then
        break
    fi
    sleep 0.5
done

# --- TEST 1: OWD PRIVATE ENFORCEMENT ---
echo "🧪 Test 1: OWD Private (Bob cannot see Alice's record)..."

# Login as Alice
api_login "$ALICE_EMAIL" "Password123!"
ALICE_TOKEN=$TOKEN
ALICE_REC=$(api_post "/api/data/salary" "{\"name\": \"Alice Salary\", \"amount\": 5000}")
ALICE_REC_ID=$(json_extract "$ALICE_REC" "id")

# Login as Bob
api_login "$BOB_EMAIL" "Password123!"
BOB_TOKEN=$TOKEN
BOB_REC=$(api_post "/api/data/salary" "{\"name\": \"Bob Salary\", \"amount\": 3000}")
BOB_REC_ID=$(json_extract "$BOB_REC" "id")

# Bob tries to get Alice's record
echo "🔍 Bob trying to access Alice's record..."
BOB_GET_ALICE=$(api_get "/api/data/salary/$ALICE_REC_ID")
if [[ "$BOB_GET_ALICE" == *"forbidden"* || "$BOB_GET_ALICE" == *"not found"* || "$BOB_GET_ALICE" == "{}" ]]; then
    echo "✅ Bob denied access to Alice's private record (as expected)"
else
    # Check if records array is empty (for queries) or 403 for direct GET
    # Our API returns 403 Forbidden for direct GET if no access
    if echo "$BOB_GET_ALICE" | grep -q "error"; then
        echo "✅ Bob access denied (error returned)"
    else
        echo "❌ Bob could see Alice's record! $BOB_GET_ALICE"
        exit 1
    fi
fi

# --- TEST 2: ROLE HIERARCHY ---
echo "🧪 Test 2: Role Hierarchy (Alice can see Bob's record)..."
# Alice (Manager) should see Bob's record because she is above in hierarchy
export TOKEN=$ALICE_TOKEN
ALICE_GET_BOB=$(api_get "/api/data/salary/$BOB_REC_ID")
ALICE_GET_BOB_NAME=$(json_extract "$ALICE_GET_BOB" "name")

if [[ "$ALICE_GET_BOB_NAME" == "Bob Salary" ]]; then
    echo "✅ Alice (Manager) can see Bob's (Assistant) record via Role Hierarchy"
else
    echo "❌ Alice could not see Bob's record: $ALICE_GET_BOB"
    exit 1
fi

# --- TEST 3: SHARING RULES ---
echo "🧪 Test 3: Sharing Rules (Public Share)..."
# Login as Admin to create sharing rule
export TOKEN=$ADMIN_TOKEN

# Create a group and add Charlie
CHARLIE_EMAIL="charlie.$TIMESTAMP@test.com"
api_post "/api/data/_System_User" "{\"username\": \"$CHARLIE_EMAIL\", \"email\": \"$CHARLIE_EMAIL\", \"password\": \"Password123!\", \"profile_id\": \"standard_user\", \"first_name\": \"Charlie\", \"last_name\": \"Accountant\"}"
api_login "$CHARLIE_EMAIL" "Password123!"
CHARLIE_TOKEN=$TOKEN
CHARLIE_ID=$USER_ID

export TOKEN=$ADMIN_TOKEN
GRP_RES=$(api_post "/api/data/_System_Group" "{\"name\": \"Accounting_$TIMESTAMP\", \"label\": \"Accounting Group $TIMESTAMP\", \"type\": \"Regular\"}")
GRP_ID=$(json_extract "$GRP_RES" "id")
if [[ -z "$GRP_ID" || "$GRP_ID" == "null" ]]; then
    echo "❌ Failed to create Group: $GRP_RES"
    exit 1
fi
echo "   Group ID: $GRP_ID"
api_post "/api/data/_System_GroupMember" "{\"group_id\": \"$GRP_ID\", \"user_id\": \"$CHARLIE_ID\"}"

# Wait for Group to be ready (Polling)
echo "   Waiting for Group propagation..."
for i in {1..10}; do
    count=$(api_post "/api/data/query" "{\"object_api_name\": \"_System_Group\", \"filters\": [{\"field\": \"id\", \"operator\": \"=\", \"value\": \"$GRP_ID\"}]}" | jq '.records | length')
    if [[ "$count" -ge 1 ]]; then
        break
    fi
    sleep 0.5
done

# Create Sharing Rule: Share all Salary records with Accounting Group
# Using Data API (criteria="true" means match all records)
api_post "/api/data/_System_SharingRule" "{
    \"name\": \"ShareWithAccounting_$TIMESTAMP\",
    \"object_api_name\": \"salary\",
    \"criteria\": \"true\",
    \"access_level\": \"Read\",
    \"share_with_group_id\": \"$GRP_ID\"
}"

# Wait for sharing rule propagation (Poll until Charlie can see Alice's record)
echo "   Waiting for Sharing Rule propagation..."
export TOKEN=$CHARLIE_TOKEN
rule_active=false
for i in {1..10}; do
    test_access=$(api_get "/api/data/salary/$ALICE_REC_ID")
    name=$(json_extract "$test_access" "name")
    if [[ "$name" == "Alice Salary" ]]; then
        rule_active=true
        break
    fi
    sleep 0.5
done

if [ "$rule_active" = false ]; then
    echo "⚠️ Warning: Sharing rule propagation timed out (might fail next step)"
fi

# Now Charlie should see Alice's record
export TOKEN=$CHARLIE_TOKEN
CHAR_GET_ALICE=$(api_get "/api/data/salary/$ALICE_REC_ID")
CHAR_GET_ALICE_NAME=$(json_extract "$CHAR_GET_ALICE" "name")

if [[ "$CHAR_GET_ALICE_NAME" == "Alice Salary" ]]; then
    echo "✅ Charlie saw Alice's record via Sharing Rule"
else
    echo "❌ Charlie could not see Alice's record: $CHAR_GET_ALICE"
    exit 1
fi

echo "✅ Suite 44 Passed!"
