#!/bin/bash
source tests/e2e/config.sh
source tests/e2e/lib/api.sh
source tests/e2e/lib/schema_helpers.sh

api_login "admin@test.com" "Admin123!"

# Ensure car exists
ensure_schema "car" "Car" "Cars"

# Try adding price field manually and print response
echo "Attempting to add price field..."
# Copied from test 43:
# add_field "car" "price" "Price" "Currency"

api_name="price"
label="Price"
type="Currency"
required="false"
schema="car"

json=$(jq -n \
    --arg api_name "$api_name" \
    --arg label "$label" \
    --arg type "$type" \
    --argjson required "$required" \
    '{api_name: $api_name, label: $label, type: $type, required: $required}')

echo "Payload: $json"

res=$(api_post "/api/metadata/objects/$schema/fields" "$json")
echo "Response: $res"
