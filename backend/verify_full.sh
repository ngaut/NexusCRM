#!/bin/bash
set -e

# 1. Login
TOKEN=$(curl -s -X POST http://localhost:3001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"Admin123!"}' | jq -r .token)

# 2. Create Account
echo "Creating Account..."
ACC_ID=$(curl -s -X POST http://localhost:3001/api/data/account \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Parent Account"}' | jq -r .record.id)

echo "Parent Account ID: $ACC_ID"

# 3. Create Child Contact
echo "Creating Child Contact..."
CONT_RES=$(curl -s -X POST http://localhost:3001/api/data/contact \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Child Contact\",\"email\":\"child@test.com\",\"account_id\":\"$ACC_ID\"}")
CONT_ID=$(echo $CONT_RES | jq -r .record.id)
echo "Child Contact ID: $CONT_ID"

# 4. Get Account Layout (Should show Contacts related list)
echo "Fetching Account Layout..."
LAYOUT_RES=$(curl -s "http://localhost:3001/api/metadata/layouts/account" \
  -H "Authorization: Bearer $TOKEN")

# Check for RELATED LISTS
# The response structure is .layout.relatedLists which is an array of objects
# We look for "Contact" or "Contacts" in it
RELATED_LISTS=$(echo $LAYOUT_RES | jq -c '.layout.relatedLists')
echo "Related Lists JSON: $RELATED_LISTS"

if [[ "$RELATED_LISTS" == *"Contacts"* ]] || [[ "$RELATED_LISTS" == *"contacts"* ]]; then
    echo "✅ SUCCESS: Contacts related list found!"
else
    echo "❌ FAILURE: Contacts related list NOT found (but query might have succeeded if empty)"
fi
