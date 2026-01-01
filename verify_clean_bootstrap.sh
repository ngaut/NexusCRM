#!/bin/bash
TARGET_FILE="backend/internal/bootstrap/system_tables.json"

echo "🔍 Verifying $TARGET_FILE for legacy Standard Object references..."

FORBIDDEN_TERMS=("Account" "Contact" "Opportunity" "Lead")
FOUND_ISSUES=0

for TERM in "${FORBIDDEN_TERMS[@]}"; do
    if grep -q "$TERM" "$TARGET_FILE"; then
        echo "❌ Found forbidden term: $TERM"
        grep -n "$TERM" "$TARGET_FILE"
        FOUND_ISSUES=1
    else
        echo "✅ No usage of: $TERM"
    fi
done

if [ $FOUND_ISSUES -eq 0 ]; then
    echo "🎉 Verification Passed: system_tables.json is clean."
    exit 0
else
    echo "⚠️ Verification Failed: Found legacy references."
    exit 1
fi
