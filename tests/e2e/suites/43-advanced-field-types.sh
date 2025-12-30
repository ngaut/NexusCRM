#!/bin/bash
set -e

# Suite 43: Advanced Field Types Lifecycle
# Tests AutoNumber, Percent, Currency.

SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"

test_cleanup() {
    echo "🧹 Cleaning up Car object..."
    delete_schema "car"
}

trap test_cleanup EXIT

# Login first
api_login "admin@test.com" "Admin123!"

# 1. Create Object 'car'
echo "🚗 Creating Car object..."
ensure_schema "car" "Car" "Cars"

# 2. Add Price (Currency)
echo "💰 Adding Price field..."
add_field "car" "price" "Price" "Currency"

# 3. Add Tax (Percent)
echo "📈 Adding Tax rate field..."
add_field "car" "tax_rate" "Tax Rate" "Percent"

# 4. Add AutoID (AutoNumber)
echo "🤖 Adding AutoID field..."
# Note: we use default_value as the format string in our implementation
add_field "car" "auto_id" "AutoID" "AutoNumber" "false" "{\"default_value\": \"CAR-{000}\"}"

# 5. Wait for cache
sleep 1

# 6. Create first Car record
echo "📝 Creating first Car record..."
REC1=$(api_post "/api/data/car" "{\"name\": \"Model S\", \"price\": 79990.0, \"tax_rate\": 0.08}")
REC1_ID=$(json_extract "$REC1" "id")
REC1_AUTO=$(json_extract "$REC1" "auto_id")

echo "✅ First Car AutoID: $REC1_AUTO"
if [[ "$REC1_AUTO" != "CAR-001" ]]; then
    echo "❌ Expected CAR-001, got $REC1_AUTO"
    exit 1
fi

# 7. Create second Car record
echo "📝 Creating second Car record..."
REC2=$(api_post "/api/data/car" "{\"name\": \"Model 3\", \"price\": 39990.0, \"tax_rate\": 0.08}")
REC2_AUTO=$(json_extract "$REC2" "auto_id")

echo "✅ Second Car AutoID: $REC2_AUTO"
if [[ "$REC2_AUTO" != "CAR-002" ]]; then
    echo "❌ Expected CAR-002, got $REC2_AUTO"
    exit 1
fi

# 8. Verify Percent and Currency persistence
echo "🔍 Verifying field values..."
GET_REC1=$(api_get "/api/data/car/$REC1_ID")
VAL_PRICE=$(json_extract "$GET_REC1" "price")
VAL_TAX=$(json_extract "$GET_REC1" "tax_rate")

echo "Price: $VAL_PRICE, Tax: $VAL_TAX"
# Check if price is within 79990 (it might be string in JSON, or float)
if [[ "$VAL_PRICE" == "79990"* ]]; then
    echo "✅ Price verified"
else
    echo "❌ Price mismatch: $VAL_PRICE"
    # Try perl if direct match fails due to trailing zeros
    perl -e "exit ($VAL_PRICE == 79990 ? 0 : 1)" || (echo "❌ Price mismatch (perl check)"; exit 1)
fi

if [[ "$VAL_TAX" == "0.08"* ]]; then
    echo "✅ Tax verified"
else
     perl -e "exit ($VAL_TAX == 0.08 ? 0 : 1)" || (echo "❌ Tax mismatch"; exit 1)
fi

echo "✅ Suite 43 Passed!"
