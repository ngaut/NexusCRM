#!/bin/bash
source tests/e2e/config.sh
source tests/e2e/lib/api.sh
source tests/e2e/lib/schema_helpers.sh

api_login "admin@test.com" "Admin123!"

echo "Deleting car schema..."
res=$(api_delete "/api/metadata/objects/car")
echo "Delete Response: $res"

echo "Checking if car exists..."
check=$(api_get "/api/metadata/objects/car")
echo "Check Response: $check"

# If check contains "api_name":"car", delete failed.
if echo "$check" | grep -q "\"api_name\":\"car\""; then
    echo "ERROR: Car schema still exists!"
else
    echo "Car schema deleted successfully."
fi
