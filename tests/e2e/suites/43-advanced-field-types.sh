#!/bin/bash
set -e

# Suite 43: Advanced Field Types Lifecycle
# Tests AutoNumber, Percent, Currency.

SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"

SUITE_NAME="Advanced Field Types"

test_cleanup() {
    echo "🧹 Cleaning up Car object..."
    delete_schema "car"
}

trap test_cleanup EXIT

run_suite() {
    section_header "$SUITE_NAME"

    # Login first
    if ! api_login "admin@test.com" "Admin123!"; then
        echo "Failed to login"
        return 1
    fi

    # 1. Create Object 'car'
    echo "🚗 Creating Car object..."
    # Ensure clean state from previous failed runs
    delete_schema "car" || true
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
    # 5. Wait for cache (Polling)
    echo "  Waiting for field 'auto_id'..."
    for i in {1..10}; do
        meta=$(api_get "/api/metadata/objects/car")
        if echo "$meta" | grep -q "\"api_name\":\"auto_id\""; then
            break
        fi
        sleep 0.5
    done

    # 6. Create first Car record
    echo "📝 Creating first Car record..."
    REC1=$(api_post "/api/data/car" "{\"name\": \"Model S\", \"price\": 79990.0, \"tax_rate\": 0.08}")
    echo "DEBUG REC1: $REC1"
    REC1_ID=$(json_extract "$REC1" "__sys_gen_id")
    REC1_AUTO=$(json_extract "$REC1" "auto_id")

    echo "✅ First Car AutoID: $REC1_AUTO"
    if [[ "$REC1_AUTO" != "CAR-001" ]]; then
        echo "❌ Expected CAR-001, got $REC1_AUTO"
        return 1
    fi

    # 7. Create second Car record
    echo "📝 Creating second Car record..."
    REC2=$(api_post "/api/data/car" "{\"name\": \"Model 3\", \"price\": 39990.0, \"tax_rate\": 0.08}")
    REC2_AUTO=$(json_extract "$REC2" "auto_id")

    echo "✅ Second Car AutoID: $REC2_AUTO"
    if [[ "$REC2_AUTO" != "CAR-002" ]]; then
        echo "❌ Expected CAR-002, got $REC2_AUTO"
        return 1
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
        # Try awk if direct match fails due to trailing zeros
        awk -v a="$VAL_PRICE" -v b="79990" 'BEGIN { exit (a == b ? 0 : 1) }' || (echo "❌ Price mismatch (awk check)"; return 1)
    fi

    if [[ "$VAL_TAX" == "0.08"* ]]; then
        echo "✅ Tax verified"
    else
         awk -v a="$VAL_TAX" -v b="0.08" 'BEGIN { exit (a == b ? 0 : 1) }' || (echo "❌ Tax mismatch"; return 1)
    fi

    echo "✅ Suite 43 Passed!"
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
