#!/bin/bash
# tests/e2e/lib/teardown.sh
# Teardown logic to clean up standard objects and apps created by setup.sh

source "$(dirname "${BASH_SOURCE[0]}")/schema_helpers.sh"

teardown_standard_objects() {
    echo "🧹 Tearing down standard objects..."
    
    # Check API availability first
    if ! api_get "/health" &>/dev/null; then
        echo "   ⚠️  Server not reachable, skipping teardown."
        return
    fi

    # Delete App
    delete_app "nexus_crm"

    # Delete Objects (Reverse order of creation to avoid constraint issues if any, 
    # though soft deletes usually handle it. Case -> Opp -> Lead -> Contact -> Account)
    delete_schema "case"
    delete_schema "opportunity"
    delete_schema "lead"
    delete_schema "contact"
    delete_schema "account"

    echo "✅ Teardown complete."
}
