#!/bin/bash
# Wait for server
sleep 5

# Login
LOGIN_RESP=$(curl -s -X POST http://localhost:3001/api/auth/login -d '{"email":"admin@test.com","password":"Admin123!"}')
TOKEN=$(echo $LOGIN_RESP | jq -r .token)

if [ "$TOKEN" == "null" ]; then
  echo "Login failed"
  echo $LOGIN_RESP
  exit 1
fi

echo "Token: $TOKEN"

# Create User
echo "Creating User..."
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  http://localhost:3001/api/auth/register \
  -d '{"name":"Test User Perms","email":"testperms@example.com","password":"User123!","profile_id":"standard_user"}'

echo ""

# Create Permission Set
echo "Creating Permission Set..."
curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  http://localhost:3001/api/auth/permission-sets \
  -d '{"name":"test_perm_set","label":"Test Perm Set","description":"Grants Modify All on Account"}'

echo ""
