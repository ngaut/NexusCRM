#!/bin/bash

# 1. Login
echo "Logging in..."
LOGIN_RES=$(curl -s -X POST http://localhost:3001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"Admin123!"}')

TOKEN=$(echo $LOGIN_RES | jq -r '.token')
USER_ID=$(echo $LOGIN_RES | jq -r '.user.id')

if [ "$TOKEN" == "null" ]; then
  echo "Login failed"
  echo $LOGIN_RES
  exit 1
fi

echo "User ID: $USER_ID"

# 2. Get an Account ID
echo "Fetching an account..."
QUERY_RES=$(curl -s -X POST http://localhost:3001/api/data/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"object_api_name":"account","limit":1}')

ACCOUNT_ID=$(echo $QUERY_RES | jq -r '.records[0].id')

if [ "$ACCOUNT_ID" == "null" ]; then
  echo "No account found, creating one..."
  # Create account if needed
   CREATE_ACC_RES=$(curl -s -X POST http://localhost:3001/api/data/account \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"Debug Account"}')
   ACCOUNT_ID=$(echo $CREATE_ACC_RES | jq -r '.record.id')
fi

echo "Account ID: $ACCOUNT_ID"

# 3. Post Comment
echo "Posting comment..."
COMMENT_PAYLOAD=$(cat <<EOF
{
  "object_api_name": "account",
  "record_id": "$ACCOUNT_ID",
  "body": "<p>Debug Comment</p>",
  "created_by_id": "$USER_ID"
}
EOF
)

echo "Payload: $COMMENT_PAYLOAD"

curl -v -X POST http://localhost:3001/api/data/_System_Comment \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$COMMENT_PAYLOAD"
