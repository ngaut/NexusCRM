#!/bin/bash
# tests/e2e/suites/34-schema-validation.sh
# Schema Validation & Assetion Tests
# Ports logic from backend/scripts/verify_assertions.sh into E2E suite

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Schema Definition Validation"
TIMESTAMP=$(date +%s)

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Define test object variable for cleanup
    local valid_obj="val_test_${TIMESTAMP}"
    
    # Cleanup function
    cleanup() {
        echo "Cleaning up..."
        api_delete "/api/metadata/objects/$valid_obj" > /dev/null 2>&1
    }
    trap cleanup EXIT
    
    test_snake_case_validation
    test_lookup_reference_validation
    test_picklist_options_validation
    test_formula_return_type_validation
}

# =========================================
# SNAKE_CASE VALIDATION
# =========================================
test_snake_case_validation() {
    echo ""
    echo "Test 34.1: API Name Snake Case Validation"
    
    # 1. Invalid CamelCase Object
    echo "  Checking CamelCase Object API Name..."
    local res=$(api_post "/api/metadata/objects" '{
        "label": "Invalid Camel",
        "plural_label": "Invalid Camels",
        "api_name": "InvalidCamelObject_'$TIMESTAMP'",
        "is_custom": true
    }')
    
    if echo "$res" | grep -qiE "must be (in )?snake_case|Invalid API name|validation error"; then
        echo "  ✓ CamelCase Object Request Rejected"
    else
        test_failed "CamelCase Object Accepted" "$res"
    fi

    # 2. Invalid Uppercase Object
    echo "  Checking Uppercase Object API Name..."
    local res2=$(api_post "/api/metadata/objects" '{
        "label": "UPPER CASE",
        "plural_label": "UPPER CASES",
        "api_name": "UPPER_CASE_OBJECT_'$TIMESTAMP'",
        "is_custom": true
    }')
    
    if echo "$res2" | grep -qiE "must be (in )?snake_case|Invalid API name|validation error"; then
        echo "  ✓ UpperCase Object Request Rejected"
    else
        test_failed "UpperCase Object Accepted" "$res2"
    fi

    # 3. Create Valid Object for Field Tests
    local valid_obj="val_test_${TIMESTAMP}"
    api_post "/api/metadata/objects" '{
        "label": "Valid Validation Test",
        "plural_label": "Valid Validation Tests",
        "api_name": "'$valid_obj'",
        "is_custom": true
    }' > /dev/null

    # 4. Invalid CamelCase Field
    echo "  Checking CamelCase Field..."
    local res3=$(api_post "/api/metadata/objects/$valid_obj/fields" '{
        "label": "Camel Field",
        "api_name": "camelField",
        "type": "Text"
    }')
    
    if echo "$res3" | grep -qiE "must be snake_case|validation failed|failed to add column"; then
        echo "  ✓ CamelCase Field Request Rejected"
        test_passed "Snake Case Validation Enforced"
    else
        test_failed "CamelCase Field Accepted" "$res3"
    fi
}

# =========================================
# LOOKUP REFERENCE VALIDATION
# =========================================
test_lookup_reference_validation() {
    echo ""
    echo "Test 34.2: Lookup Reference Validation"
    local valid_obj="val_test_${TIMESTAMP}"

    # 1. Lookup without Reference
    echo "  Checking Lookup without reference_to..."
    local res=$(api_post "/api/metadata/objects/$valid_obj/fields" '{
        "label": "Bad Lookup",
        "api_name": "bad_lookup",
        "type": "Lookup"
    }')
    
    if echo "$res" | grep -qiE "reference_to is required|validation failed"; then
        echo "  ✓ Missing Reference Rejected"
        test_passed "Lookup Reference Validation Enforced"
    else
        test_failed "Lookup without reference Accepted" "$res"
    fi
}

# =========================================
# PICKLIST OPTIONS VALIDATION
# =========================================
test_picklist_options_validation() {
    echo ""
    echo "Test 34.3: Picklist Options Validation"
    local valid_obj="val_test_${TIMESTAMP}"

    # 1. Picklist without Options
    echo "  Checking Picklist without options..."
    local res=$(api_post "/api/metadata/objects/$valid_obj/fields" '{
        "label": "Bad Picklist",
        "api_name": "bad_picklist",
        "type": "Picklist"
    }')
    
    if echo "$res" | grep -qiE "options are required|validation failed"; then
        echo "  ✓ Missing Options Rejected"
        test_passed "Picklist Options Validation Enforced"
    else
        test_failed "Picklist without options Accepted" "$res"
    fi
}

# =========================================
# FORMULA VALIDATION
# =========================================
test_formula_return_type_validation() {
    echo ""
    echo "Test 34.4: Formula Return Type Validation"
    local valid_obj="val_test_${TIMESTAMP}"

    # 1. Formula without Return Type
    echo "  Checking Formula without return_type..."
    local res=$(api_post "/api/metadata/objects/$valid_obj/fields" '{
        "label": "Bad Formula",
        "api_name": "bad_formula",
        "type": "Formula",
        "formula": "2+2"
    }')
    
    if echo "$res" | grep -qiE "return_type is required|validation failed"; then
        echo "  ✓ Missing Return Type Rejected"
        test_passed "Formula Return Type Validation Enforced"
    else
        test_failed "Formula without return_type Accepted" "$res"
    fi
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
