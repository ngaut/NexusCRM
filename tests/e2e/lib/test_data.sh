#!/bin/bash
# tests/e2e/lib/test_data.sh
# Factory functions for creating test data with required fields
# Ensures consistent test data across all E2E suites

# Get directory of this file
TEST_DATA_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source dependencies if not loaded
if [ -z "$BASE_URL" ]; then
    source "$TEST_DATA_DIR/../config.sh"
fi
if ! type api_post &>/dev/null; then
    source "$TEST_DATA_DIR/api.sh"
fi

# =========================================
# STANDARD OBJECT FIELD REQUIREMENTS
# =========================================

# These constants document required fields for standard objects
# Used by factory functions to ensure test data is valid

# Note: Required fields are documented in each factory function below

# =========================================
# FACTORY FUNCTIONS
# =========================================

# Create a Lead record
# Usage: lead_id=$(create_lead "Lead Name" "email@test.com" '{"extra": "fields"}')
# Returns: record ID or empty string on failure
create_lead() {
    local name="${1:-Test Lead $RANDOM}"
    local email="${2:-lead.$RANDOM@test.com}"
    local extra="${3:-}"
    
    # Note: Lead requires name, email, AND company
    local json=$(jq -n \
        --arg name "$name" \
        --arg email "$email" \
        '{name: $name, email: $email, company: "Test Company", status: "New"}')
    
    # Merge extra fields if provided
    if [ -n "$extra" ]; then
        json=$(echo "$json" | jq --argjson extra "$extra" '. + $extra')
    fi
    
    local res=$(api_post "/api/data/lead" "$json")
    json_extract "$res" "id"
}

# Create an Account record
# Usage: account_id=$(create_account "Account Name" '{"extra": "fields"}')
create_account() {
    local name="${1:-Test Account $RANDOM}"
    local extra="${2:-}"
    
    local json=$(jq -n --arg name "$name" '{name: $name}')
    
    if [ -n "$extra" ]; then
        json=$(echo "$json" | jq --argjson extra "$extra" '. + $extra')
    fi
    
    local res=$(api_post "/api/data/account" "$json")
    json_extract "$res" "id"
}

# Create a Contact record
# Usage: contact_id=$(create_contact "Contact Name" "account_id" '{"extra": "fields"}')
create_contact() {
    local name="${1:-Test Contact $RANDOM}"
    local account_id="$2"
    local extra="${3:-}"
    
    local json=$(jq -n --arg name "$name" '{name: $name}')
    
    if [ -n "$account_id" ]; then
        json=$(echo "$json" | jq --arg acc "$account_id" '. + {account_id: $acc}')
    fi
    
    if [ -n "$extra" ]; then
        json=$(echo "$json" | jq --argjson extra "$extra" '. + $extra')
    fi
    
    local res=$(api_post "/api/data/contact" "$json")
    json_extract "$res" "id"
}

# Create an Opportunity record
# Usage: opp_id=$(create_opportunity "Opp Name" "account_id" '{"extra": "fields"}')
create_opportunity() {
    local name="${1:-Test Opportunity $RANDOM}"
    local account_id="$2"
    local extra="${3:-}"
    
    # Note: stage_name is required, not stage
    local json=$(jq -n \
        --arg name "$name" \
        '{name: $name, stage_name: "Prospecting"}')
    
    if [ -n "$account_id" ]; then
        json=$(echo "$json" | jq --arg acc "$account_id" '. + {account_id: $acc}')
    fi
    
    if [ -n "$extra" ]; then
        json=$(echo "$json" | jq --argjson extra "$extra" '. + $extra')
    fi
    
    local res=$(api_post "/api/data/opportunity" "$json")
    json_extract "$res" "id"
}

# =========================================
# BULK CREATE HELPERS
# =========================================

# Create multiple records and return IDs as array
# Usage: ids=($(bulk_create_leads 5 "prefix"))
bulk_create_leads() {
    local count="${1:-10}"
    local prefix="${2:-BulkLead}"
    local ids=()
    
    for i in $(seq 1 "$count"); do
        local id=$(create_lead "$prefix $i" "${prefix,,}$i.$RANDOM@test.com")
        if [ -n "$id" ]; then
            ids+=("$id")
        fi
    done
    
    echo "${ids[@]}"
}

bulk_create_accounts() {
    local count="${1:-10}"
    local prefix="${2:-BulkAccount}"
    local ids=()
    
    for i in $(seq 1 "$count"); do
        local id=$(create_account "$prefix $i")
        if [ -n "$id" ]; then
            ids+=("$id")
        fi
    done
    
    echo "${ids[@]}"
}

# =========================================
# CLEANUP HELPERS
# =========================================

# Delete multiple records by ID
# Usage: cleanup_records "lead" "${LEAD_IDS[@]}"
cleanup_records() {
    local object="$1"
    shift
    local ids=("$@")
    local deleted=0
    
    for id in "${ids[@]}"; do
        if [ -n "$id" ]; then
            api_delete "/api/data/$object/$id" > /dev/null 2>&1
            deleted=$((deleted + 1))
        fi
    done
    
    echo "  ✓ Deleted $deleted $object records"
}

# =========================================
# DATA GENERATION HELPERS
# =========================================

# Generate unique email
unique_email() {
    local prefix="${1:-user}"
    echo "${prefix}.${RANDOM}${RANDOM}@test.com"
}

# Generate unique phone (10 digits)
unique_phone() {
    printf "555%07d" $RANDOM
}

# Generate timestamp-based unique name
unique_name() {
    local prefix="${1:-Record}"
    echo "$prefix $(date +%s)$RANDOM"
}
