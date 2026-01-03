#!/bin/bash
# tests/e2e/runner.sh
# E2E Test Suite Runner

set +e  # Don't exit on error - we want to see all test results

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source dependencies
source "$SCRIPT_DIR/config.sh"
source "$SCRIPT_DIR/lib/helpers.sh"
source "$SCRIPT_DIR/lib/teardown.sh"

# Default configs
CLEANUP_ON_EXIT=false

# Initialize counters
export TOTAL_PASSED=0
export TOTAL_FAILED=0

AVAILABLE_SUITES=()

# Dynamic Suite Discovery
# Finds all .sh files in suites/ starting with a number
discover_suites() {
    local suite_dir="$SCRIPT_DIR/suites"
    if [ -d "$suite_dir" ]; then
        # Use find to get files, sort them, and process
        while IFS= read -r file; do
            # Extract filename without extension
            basename=$(basename "$file" .sh)
            AVAILABLE_SUITES+=("$basename")
        done < <(find "$suite_dir" -maxdepth 1 -name "[0-9]*.sh" | sort)
    else
        echo "Error: Suites directory not found at $suite_dir"
        exit 1
    fi
}

# Run discovery
discover_suites

# Print usage
usage() {
    echo "Usage: $0 [OPTIONS] [SUITES...]"
    echo ""
    echo "Options:"
    echo "  --list, -l       List available test suites"
    echo "  --cleanup        Run teardown after tests"
    echo "  --help, -h       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all test suites"
    echo "  $0 auth metadata      # Run specific suites"
    echo "  $0 01 02             # Run suites by number"
    exit 0
}

# List available suites
list_suites() {
    echo "Available Test Suites:"
    echo ""
    for suite in "${AVAILABLE_SUITES[@]}"; do
        suite_file="$SCRIPT_DIR/suites/$suite.sh"
        if [ -f "$suite_file" ]; then
            suite_name=$(grep 'SUITE_NAME=' "$suite_file" | head -1 | cut -d'"' -f2)
            echo "  $suite - $suite_name"
        fi
    done
    exit 0
}

# Run a single test suite
run_test_suite() {
    local suite_num="$1"
    local suite_file="$SCRIPT_DIR/suites/$suite_num.sh"
    
    if [ ! -f "$suite_file" ]; then
        echo -e "${RED}✗ Suite not found: $suite_num${NC}"
        return 1
    fi
    
    # Run the suite
    source "$suite_file"
    run_suite
}

# Main execution
echo ""
echo "═══════════════════════════════════════════════════════════"
echo "   NexusCRM E2E Test Suite"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "Testing from a real user's perspective:"
echo "  - Authentication & Session Management"
echo "  - Metadata Discovery"
echo "  - Record CRUD Operations"
echo "  - Search & Query"
echo "  - Formula Evaluation"
echo "  - Recycle Bin Operations"
echo ""
echo "═══════════════════════════════════════════════════════════"
echo ""

# Parse arguments
SUITES_TO_RUN=()

if [ $# -eq 0 ]; then
    # No arguments - run all suites
    SUITES_TO_RUN=("${AVAILABLE_SUITES[@]}")
else
    # Parse arguments
    for arg in "$@"; do
        case $arg in
            --cleanup)
                CLEANUP_ON_EXIT=true
                ;;
            --list|-l)
                list_suites
                ;;
            --help|-h)
                usage
                ;;
            *)
                # Check if it's a number (suite number)
                if [[ $arg =~ ^[0-9]+$ ]]; then
                    # Find suite by number
                    suite_num=$(printf "%02d" $((10#$arg)))
                    for suite in "${AVAILABLE_SUITES[@]}"; do
                        if [[ $suite == $suite_num-* ]]; then
                            SUITES_TO_RUN+=("$suite")
                            break
                        fi
                    done
                else
                    # Try to find suite by name
                    found=false
                    for suite in "${AVAILABLE_SUITES[@]}"; do
                        if [[ $suite == *"$arg"* ]]; then
                            SUITES_TO_RUN+=("$suite")
                            found=true
                            break
                        fi
                    done
                    if [ "$found" = false ]; then
                        echo -e "${YELLOW}Warning: Suite not found: $arg${NC}"
                    fi
                fi
                ;;
        esac
    done
fi


# Global Setup: Ensure standard objects exist
if [ ${#SUITES_TO_RUN[@]} -gt 0 ]; then
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo "   Global Setup"
    echo "═══════════════════════════════════════════════════════════"
    
    source "$SCRIPT_DIR/lib/setup.sh"
    
    # We need a token for setup
    if api_login; then
        ensure_standard_objects_exist
    else
        echo -e "${RED}✗ Global Login Failed. Aborting.${NC}"
        exit 1
    fi
    echo ""
fi

# Run selected suites
for suite in "${SUITES_TO_RUN[@]}"; do
    run_test_suite "$suite"
done

# Print summary
echo ""
echo "═══════════════════════════════════════════════════════════"
echo "   Test Results Summary"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo -e "Total Tests: $(($TOTAL_PASSED + $TOTAL_FAILED))"
echo -e "Passed: ${GREEN}$TOTAL_PASSED${NC}"
echo -e "Failed: ${RED}$TOTAL_FAILED${NC}"

if [ $TOTAL_FAILED -eq 0 ]; then
    PASS_RATE="100%"
else
    PASS_RATE=$(awk "BEGIN {printf \"%.1f%%\", ($TOTAL_PASSED/($TOTAL_PASSED+$TOTAL_FAILED))*100}")
fi
echo -e "Pass Rate: ${GREEN}$PASS_RATE${NC}"

echo ""
echo "═══════════════════════════════════════════════════════════"

# Teardown
if [ "$CLEANUP_ON_EXIT" = true ]; then
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo "   Global Teardown"
    echo "═══════════════════════════════════════════════════════════"
    if api_login; then
        teardown_standard_objects
    fi
fi

if [ $TOTAL_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
    echo ""
    echo "NexusCRM Backend: Production Ready ✅"
    echo "  ✓ All user workflows tested"
    echo "  ✓ Authentication & authorization working"
    echo "  ✓ CRUD operations functional"
    echo "  ✓ Search & query working"
    echo "  ✓ Formula engine operational"
    echo ""
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo ""
    echo "Please review failed tests above and fix issues."
    echo ""
    exit 1
fi
