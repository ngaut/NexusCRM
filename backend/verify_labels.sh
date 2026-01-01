#!/bin/bash
TOKEN=$(curl -s -X POST http://localhost:3001/api/auth/login -H "Content-Type: application/json" -d '{"email":"admin@test.com", "password":"Admin123!"}' | jq -r '.token')

echo "---------------------------------------------------"
echo "Checking _System_User metadata..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/metadata/objects/_System_User | jq '.schema.fields[] | select(.api_name == "username" or .api_name == "email" or .api_name == "created_date" or .api_name == "id") | {api_name, label}'

echo "---------------------------------------------------"
echo "Checking _System_Role metadata..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/metadata/objects/_System_Role | jq '.schema.fields[] | select(.api_name == "name" or .api_name == "parent_role_id" or .api_name == "owner_id") | {api_name, label}'

echo "---------------------------------------------------"
echo "Checking _System_Profile metadata..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/metadata/objects/_System_Profile | jq '.schema.fields[] | select(.api_name == "name" or .api_name == "is_system" or .api_name == "description") | {api_name, label}'

echo "---------------------------------------------------"
echo "Checking _System_Object metadata..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/metadata/objects/_System_Object | jq '.schema.fields[] | select(.api_name == "api_name" or .api_name == "plural_label" or .api_name == "table_type") | {api_name, label}'

echo "---------------------------------------------------"
echo "Checking _System_Field metadata..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/metadata/objects/_System_Field | jq '.schema.fields[] | select(.api_name == "api_name" or .api_name == "required" or .api_name == "is_name_field") | {api_name, label}'

echo "---------------------------------------------------"
echo "Checking _System_Table metadata..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/metadata/objects/_System_Table | jq '.schema.fields[] | select(.api_name == "table_name" or .api_name == "description" or .api_name == "is_managed") | {api_name, label}'
