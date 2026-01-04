#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/api.sh"

echo "Logging in..."
if ! api_login "admin@test.com" "Admin123!"; then
    echo "Login failed"
    exit 1
fi

echo "--- ACCOUNT METADATA ---"
api_get "/api/metadata/objects/account"

echo -e "\n\n--- ACCOUNT LAYOUT ---"
api_get "/api/metadata/layouts/account"
