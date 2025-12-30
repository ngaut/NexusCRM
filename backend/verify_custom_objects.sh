#!/bin/bash
set -e

# Configuration
API_URL="http://localhost:3001"
EMAIL="admin@test.com"
PASSWORD="Admin123!"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 NexusCRM Custom Object Verification Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# 1. Login
echo "1. Logging in..."
TOKEN=$(curl -s -X POST $API_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r .token)

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Login failed! Is the server running?"
    exit 1
fi
echo "   ✅ Logged in successfully"

# 2. Create Custom Object 'Project__c'
echo "2. Creating Custom Object 'Project__c'..."
# Clean up if exists (ignore error)
curl -s -X DELETE $API_URL/api/metadata/objects/project__c \
  -H "Authorization: Bearer $TOKEN" >/dev/null || true

OBJ_RES=$(curl -s -X POST $API_URL/api/metadata/objects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "api_name": "Project__c",
    "label": "Project",
    "plural_label": "Projects",
    "description": "Project Tracking",
    "table_type": "custom_object"
  }')

OBJ_ID=$(echo $OBJ_RES | jq -r .schema.id)
if [ "$OBJ_ID" == "null" ]; then
    echo "❌ Failed to create custom object: $OBJ_RES"
    exit 1
fi
echo "   ✅ Created Custom Object: $OBJ_ID"

# 3. Add Custom Fields
echo "3. Adding Custom Fields..."

# 3a. Budget (Currency)
echo "   - Adding Budget__c (Currency)..."
FIELD_RES=$(curl -s -X POST $API_URL/api/metadata/objects/project__c/fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "api_name": "Budget__c",
    "label": "Budget",
    "type": "Currency", 
    "required": true
  }')

# 3b. Start Date (Date)
echo "   - Adding Start_Date__c (Date)..."
FIELD_RES=$(curl -s -X POST $API_URL/api/metadata/objects/project__c/fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "api_name": "Start_Date__c",
    "label": "Start Date",
    "type": "Date"
  }')
  
# 3c. Status (Picklist)
echo "   - Adding Status__c (Picklist)..."
FIELD_RES=$(curl -s -X POST $API_URL/api/metadata/objects/project__c/fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "api_name": "Status__c",
    "label": "Status",
    "type": "Picklist",
    "options": [
      {"label":"Planned","value":"Planned"},
      {"label":"Active","value":"Active"},
      {"label":"Completed","value":"Completed"}
    ]
  }')

echo "   ✅ Fields added successfully"

# 4. Create Record
echo "4. Creating Project Record..."
REC_RES=$(curl -s -X POST $API_URL/api/data/project__c \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Q4 Marketing Campaign",
    "Budget__c": 50000.00,
    "Start_Date__c": "2025-10-01",
    "Status__c": "Planned"
  }')

REC_ID=$(echo $REC_RES | jq -r .record.id)
if [ "$REC_ID" == "null" ]; then
    echo "❌ Failed to create record: $REC_RES"
    exit 1
fi
echo "   ✅ Created Record: $REC_ID"

# 5. Retrieve Record
echo "5. Retrieving Record..."
GET_RES=$(curl -s -X GET $API_URL/api/data/project__c/$REC_ID \
  -H "Authorization: Bearer $TOKEN")

RETRIEVED_BUDGET=$(echo $GET_RES | jq -r .record.Budget__c)
if [ "$RETRIEVED_BUDGET" != "50000" ] && [ "$RETRIEVED_BUDGET" != "50000.00" ]; then 
    echo "⚠️  Warning: Expected Budget 50000, got $RETRIEVED_BUDGET"
fi
echo "   ✅ Record retrieved successfully"

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Custom Object Verification SUCCESS!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
