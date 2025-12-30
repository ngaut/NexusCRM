#!/bin/bash

# Login to get token
LOGIN_RESP=$(curl -s -X POST http://localhost:3001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@test.com", "password": "Admin123!"}')

TOKEN=$(echo $LOGIN_RESP | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Login failed"
  exit 1
fi

echo "Got Token: ${TOKEN:0:10}..."

# Test 1: Invalid Name (CamelCase)
echo "--------------------------------"
echo "Test 1: Invalid Name (CamelCase)"
RESP=$(curl -s -X POST http://localhost:3001/api/metadata/objects/contact/fields \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"api_name": "BadName", "label": "Bad Name", "type": "Text"}')
echo "Response: $RESP"
echo $RESP | grep "snake_case" && echo "✅ Passed" || echo "❌ Failed"

# Test 2: Invalid Lookup (No Ref)
echo "--------------------------------"
echo "Test 2: Invalid Lookup (No Ref)"
RESP=$(curl -s -X POST http://localhost:3001/api/metadata/objects/contact/fields \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"api_name": "bad_lookup", "label": "Bad Lookup", "type": "Lookup"}')
echo "Response: $RESP"
echo $RESP | grep "reference_to" && echo "✅ Passed" || echo "❌ Failed"

# Test 3: Invalid Picklist (No Options)
echo "--------------------------------"
echo "Test 3: Invalid Picklist (No Options)"
RESP=$(curl -s -X POST http://localhost:3001/api/metadata/objects/contact/fields \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"api_name": "bad_picklist", "label": "Bad Picklist", "type": "Picklist"}')
echo "Response: $RESP"
echo $RESP | grep "option" && echo "✅ Passed" || echo "❌ Failed"

echo "--------------------------------"
echo "Done"
