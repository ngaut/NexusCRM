#!/bin/bash
set -e

# Configuration
API_URL="http://localhost:3001"
EMAIL="admin@test.com"
PASSWORD="Admin123!"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🤖 NexusCRM AI Assistant Verification"
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

# 2. Cleanup & Setup (Define Contact Object)
echo "2. Setting up Contact Object (Metadata Driven)..."
# Check if object exists, if not create it
OBJECT_CHECK=$(curl -s -H "Authorization: Bearer $TOKEN" $API_URL/api/metadata/objects/contact)
if [[ $OBJECT_CHECK == *"not found"* ]] || [[ $OBJECT_CHECK == *"error"* ]]; then
    echo "   Creating 'contact' object definition..."
    curl -s -X POST $API_URL/api/metadata/objects \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "api_name": "contact",
        "label": "Contact",
        "plural_label": "Contacts",
        "description": "Contact records",
        "table_type": "custom_object"
      }'
      
    echo "   Adding 'email' field..."
    curl -s -X POST $API_URL/api/metadata/objects/contact/fields \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "api_name": "email",
        "label": "Email",
        "type": "email",
        "required": true
      }'
      
    echo "   ✅ Object setup complete"
else
    echo "   ✅ Contact object already exists"
fi

# Cleanup previous test data
SEARCH_RES=$(curl -s -X POST $API_URL/api/data/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "object_api_name": "contact",
    "filter": "email = \u0027ai_test_user@example.com\u0027"
  }')

CONTACT_ID=$(echo $SEARCH_RES | jq -r '.records[0].id')
if [ "$CONTACT_ID" != "null" ] && [ "$CONTACT_ID" != "" ]; then
    echo "   Found existing contact ($CONTACT_ID). Deleting..."
    curl -s -X DELETE $API_URL/api/data/contact/$CONTACT_ID \
      -H "Authorization: Bearer $TOKEN" >/dev/null
    echo "   ✅ Cleanup complete"
else
    echo "   ✅ No cleanup needed"
fi

# 3. Send Request to AI Agent
echo "3. Sending Request to AI Agent..."
echo "   Prompt: 'Create a new contact named AI Test User with email ai_test_user@example.com'"

# Use curl -N for streaming, store output in a temporary file
# We use a simple loop to read the stream or just capture it all.
# Since the agent might take a few seconds and multiple steps (thinking, tool call, result, final answer),
# we capture the output.

curl -s -N -X POST $API_URL/api/agent/chat/stream \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {
        "role": "user", 
        "content": "Create a new contact named AI Test User with email ai_test_user@example.com"
      }
    ]
  }' > ai_response.txt &

PID=$!
echo "   Waiting for agent to process (60s)..."
sleep 60
kill $PID 2>/dev/null || true

echo "   ✅ Agent processing complete (or timed out)"
echo "   Analyze response events:"
grep -o '"type":"[^"]*"' ai_response.txt | sort | uniq -c

# 4. Verify Record Creation
echo "4. Verifying Contact Creation..."
# Allow a moment for async processing if any
sleep 2

VERIFY_RES=$(curl -s -X POST $API_URL/api/data/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "object_api_name": "contact",
    "filter": "email = \u0027ai_test_user@example.com\u0027"
  }')

NEW_ID=$(echo $VERIFY_RES | jq -r '.records[0].id')
NEW_NAME=$(echo $VERIFY_RES | jq -r '.records[0].name')

if [ "$NEW_ID" != "null" ] && [ "$NEW_ID" != "" ]; then
    echo "   ✅ SUCCESS: Contact created!"
    echo "      ID: $NEW_ID"
    echo "      Name: $NEW_NAME"
else
    echo "   ❌ FAILURE: Contact not found."
    echo "      Query Result: $VERIFY_RES"
    echo "      AI Response Dump:"
    cat ai_response.txt
    exit 1
fi

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ AI Assistant Verification SUCCESS!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
