#!/bin/bash
API_URL="http://localhost:3001"
CONTACT_ID="377e71e8-c075-4fec-aae3-1aa3dd8ab79f"

# Login
TOKEN=$(curl -s -X POST $API_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"Admin123!"}' | jq -r .token)

# Get Record
echo "Fetching Contact $CONTACT_ID..."
curl -s -X GET $API_URL/api/data/contact/$CONTACT_ID \
  -H "Authorization: Bearer $TOKEN" > record_detail.json

# Check HTTP Status (simulated)
if grep -q "id" record_detail.json; then
    echo "✅ Success: Record found."
    cat record_detail.json
else
    echo "❌ Failure: Record not found."
    cat record_detail.json
fi
