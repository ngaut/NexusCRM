#!/bin/bash
# tests/e2e/suites/06-formulas.sh
# Formula Engine Tests

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"

SUITE_NAME="Formula Engine"

run_suite() {
    section_header "$SUITE_NAME"
    
    test_simple_formula
    test_formula_with_context
    test_available_functions
    test_validate_syntax
    test_complex_formulas
    test_syntax_errors
}

test_simple_formula() {
    echo "Test 6.1: Evaluate Simple Formula"
    
    local response=$(api_post "/api/formula/evaluate" '{"expression": "2 + 2", "context": {}}')
    if echo "$response" | grep -qE '"result"|"value"' && echo "$response" | grep -q '4'; then
        test_passed "POST /api/formula/evaluate calculates 2+2=4"
    else
        test_failed "POST /api/formula/evaluate (2+2)" "$response"
    fi
}

test_formula_with_context() {
    echo ""
    echo "Test 6.2: Evaluate Formula with Context"
    
    local response=$(api_post "/api/formula/evaluate" '{"expression": "AnnualRevenue * 0.1", "context": {"AnnualRevenue": 1000000}}')
    if echo "$response" | grep -qE '"result"|"value"'; then
        test_passed "POST /api/formula/evaluate with context works"
    else
        test_failed "POST /api/formula/evaluate with context" "$response"
    fi
}

test_available_functions() {
    echo ""
    echo "Test 6.3: Get Available Formula Functions"
    
    local response=$(api_get "/api/formula/functions")
    if echo "$response" | grep -qE '"functions"|"LEN"|"UPPER"'; then
        local func_count=$(echo "$response" | grep -o '"name"' | wc -l)
        test_passed "GET /api/formula/functions returns function list ($func_count functions)"
    else
        test_failed "GET /api/formula/functions" "$response"
    fi
}

test_validate_syntax() {
    echo ""
    echo "Test 6.4: Validate Formula Syntax"
    
    local response=$(api_post "/api/formula/validate" '{"expression": "LEN(Name)"}')
    if echo "$response" | grep -qE '"valid"|"isValid"'; then
        test_passed "POST /api/formula/validate validates formulas"
    else
        test_failed "POST /api/formula/validate" "$response"
    fi
}

test_complex_formulas() {
    echo ""
    echo "Test 6.5: Complex Nested Formula"
    
    local response=$(api_post "/api/formula/evaluate" '{"expression": "IF(LEN(\"test\") > 3, UPPER(\"success\"), LOWER(\"FAIL\"))", "context": {}}')
    if echo "$response" | grep -qE '"result"|"value"'; then
        test_passed "Complex nested formula (IF + LEN + UPPER) evaluates"
    else
        test_failed "Complex nested formula" "$response"
    fi
}

test_syntax_errors() {
    echo ""
    echo "Test 6.6: Formula Syntax Error Handling"
    
    local response=$(api_post "/api/formula/evaluate" '{"expression": "INVALID_FUNC(2 + +)", "context": {}}')
    if echo "$response" | grep -qE "error|invalid|syntax"; then
        test_passed "Formula engine properly handles syntax errors"
    else
        test_failed "Formula syntax error handling" "$response"
    fi
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
