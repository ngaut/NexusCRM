#!/bin/bash
# tests/e2e/lib/debug_fix.sh

# Ensure we are in the right directory or source correctly
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$DIR/api.sh"

echo "Logging in..."
api_login "admin@test.com" "Admin123!"

echo "--- PATCH FIELD TEST ---"
# Try to update the 'type' field options
resp=$(api_patch "/api/metadata/objects/account/fields/type" '{"api_name": "type", "label": "Type", "type": "Picklist", "options": ["Option A", "Option B"]}')
echo "Response: $resp"

echo -e "\n\n--- POST LAYOUT TEST ---"
# Delete existing layout
layout_resp=$(api_get "/api/metadata/layouts/account")
layout_id=$(echo "$layout_resp" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')

if [ ! -z "$layout_id" ]; then
    echo "Deleting layout $layout_id"
    api_delete "/api/metadata/layouts/$layout_id"
fi

echo "Creating Layout..."
# Create a new layout with a distinct label
resp=$(api_post "/api/metadata/layouts" '{
            "object_api_name": "account",
            "label": "Account Layout MANUAL",
            "type": "Detail",
            "is_default": true,
            "sections": [
                {
                    "id": "info",
                    "label": "Information",
                    "columns": 2,
                    "fields": ["name"]
                }
            ],
            "related_lists": [
                {
                    "id": "rl_contacts",
                    "label": "Contacts",
                    "object_api_name": "contact",
                    "lookup_field": "account_id",
                    "fields": ["first_name"]
                }
            ]
        }')
echo "Response: $resp"

echo -e "\n\n--- VERIFY ---"
echo "Account Field Type:"
api_get "/api/metadata/objects/account" | grep "Option A"
echo "Account Layout Label:"
api_get "/api/metadata/layouts/account" | grep "Account Layout MANUAL"
