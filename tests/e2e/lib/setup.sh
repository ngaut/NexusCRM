#!/bin/bash
# tests/e2e/lib/setup.sh
# Ensures test objects exist before running tests
# Note: These objects (Account, Contact, Lead, etc.) are created as CUSTOM objects
# for testing purposes, replacing the legacy "Standard Objects".

source "$(dirname "${BASH_SOURCE[0]}")/api.sh"

ensure_field() {
    local obj_name="$1"
    local field_name="$2"
    local field_data="$3"

    # Strategy: Try to create. If it fails (likely exists), then update.
    # This avoids brittle grep checks on large JSON responses.
    
    echo "    Ensuring field $field_name on $obj_name..."
    local resp=$(api_post "/api/metadata/objects/$obj_name/fields" "$field_data")
    
    # Check if creation failed (assuming failure means it exists)
    if echo "$resp" | grep -q -i "error"; then
        echo "      Creation failed (likely exists), updating..."
        api_patch "/api/metadata/objects/$obj_name/fields/$field_name" "$field_data" > /dev/null
    fi
}

ensure_layout() {
    local obj_name="$1"
    local layout_data="$2"

    echo "    Ensuring Layout for $obj_name..."
    
    # 1. Find existing REAL layout ID (not default_)
    # We get the first layout that doesn't start with "default_"
    local layout_resp=$(api_get "/api/metadata/layouts/$obj_name")
    local existing_id=$(echo "$layout_resp" | grep -o '"id":"[^"]*' | sed 's/"id":"//' | grep -v "^default_" | head -1)

    local target_id=""
    
    if [ ! -z "$existing_id" ]; then
        echo "    Found existing layout $existing_id. Updating..."
        target_id="$existing_id"
    else
        echo "    No existing custom layout found. Creating new..."
        target_id=$(uuidgen | tr '[:upper:]' '[:lower:]')
    fi
    
    # Inject ID into JSON data
    # Assuming layout_data starts with {
    # We replace the starting { with { "id": "TARGET_ID", 
    local data_with_id=$(echo "$layout_data" | sed "s/{/{\"id\": \"$target_id\", /")

    local resp=$(api_post "/api/metadata/layouts" "$data_with_id")
    
    if echo "$resp" | grep -q -i "error"; then
        echo "    ❌ Layout upsert failed: $resp"
    fi
}

ensure_test_objects() {
    echo "⚡️ Ensuring E2E test objects exist..."

    # ==========================
    # ACCOUNT
    # ==========================
    if ! api_get "/api/metadata/objects/account" | grep -q '"api_name":"account"'; then
        echo "   Creating Account object..."
        api_post "/api/metadata/objects" '{
            "label": "Account",
            "plural_label": "Accounts",
            "api_name": "account",
            "description": "Account (Test Object)",
            "is_custom": true,
            "theme_color": "#3b82f6",
            "list_fields": ["name", "industry", "website", "phone"]
        }' > /dev/null
    fi
    
    ensure_field "account" "name" '{"api_name": "name", "label": "Account Name", "type": "Text", "required": true, "is_name_field": true}'
    ensure_field "account" "industry" '{"api_name": "industry", "label": "Industry", "type": "Text"}'
    ensure_field "account" "type" '{"api_name": "type", "label": "Type", "type": "Picklist", "options": ["Prospect", "Customer - Direct", "Customer - Channel", "Partner", "Competitor"]}'
    ensure_field "account" "phone" '{"api_name": "phone", "label": "Phone", "type": "Phone"}'
    ensure_field "account" "website" '{"api_name": "website", "label": "Website", "type": "Url"}'

    # Update list fields and theme color
    api_patch "/api/metadata/objects/account" '{"list_fields": ["name", "industry", "type", "website", "phone"], "theme_color": "#3b82f6"}' > /dev/null


    # ==========================
    # CONTACT
    # ==========================
    if ! api_get "/api/metadata/objects/contact" | grep -q '"api_name":"contact"'; then
        echo "   Creating Contact object..."
        api_post "/api/metadata/objects" '{
            "label": "Contact",
            "plural_label": "Contacts",
            "api_name": "contact",
            "description": "Contact (Test Object)",
            "is_custom": true,
            "list_fields": ["first_name", "last_name", "email", "phone"]
        }' > /dev/null
    fi

    ensure_field "contact" "first_name" '{"api_name": "first_name", "label": "First Name", "type": "Text"}'
    ensure_field "contact" "last_name" '{"api_name": "last_name", "label": "Last Name", "type": "Text", "required": true, "is_name_field": true}'
    ensure_field "contact" "email" '{"api_name": "email", "label": "Email", "type": "Email"}'
    ensure_field "contact" "phone" '{"api_name": "phone", "label": "Phone", "type": "Phone"}'
    ensure_field "contact" "title" '{"api_name": "title", "label": "Title", "type": "Text"}'
    ensure_field "contact" "account_id" '{"api_name": "account_id", "label": "Account", "type": "Lookup", "reference_to": ["account"]}'

    api_patch "/api/metadata/objects/contact" '{"list_fields": ["first_name", "last_name", "email", "phone", "account_id"]}' > /dev/null


    # ==========================
    # LEAD
    # ==========================
    if ! api_get "/api/metadata/objects/lead" | grep -q '"api_name":"lead"'; then
        echo "   Creating Lead object..."
        api_post "/api/metadata/objects" '{
            "label": "Lead",
            "plural_label": "Leads",
            "api_name": "lead",
            "description": "Lead (Test Object)",
            "is_custom": true,
            "list_fields": ["name", "company", "email"]
        }' > /dev/null
    fi

    ensure_field "lead" "company" '{"api_name": "company", "label": "Company", "type": "Text", "required": true}'
    ensure_field "lead" "name" '{"api_name": "name", "label": "Name", "type": "Text", "required": true, "is_name_field": true}'
    ensure_field "lead" "email" '{"api_name": "email", "label": "Email", "type": "Email"}'
    ensure_field "lead" "status" '{"api_name": "status", "label": "Status", "type": "Picklist", "options": ["Open", "Contacted", "Qualified", "Unqualified"]}'

    api_patch "/api/metadata/objects/lead" '{"list_fields": ["name", "company", "status", "email"]}' > /dev/null


    # ==========================
    # OPPORTUNITY
    # ==========================
    if ! api_get "/api/metadata/objects/opportunity" | grep -q '"api_name":"opportunity"'; then
        echo "   Creating Opportunity object..."
        api_post "/api/metadata/objects" '{
            "label": "Opportunity",
            "plural_label": "Opportunities",
            "api_name": "opportunity",
            "description": "Opportunity (Test Object)",
            "is_custom": true,
            "list_fields": ["name", "amount", "close_date"]
        }' > /dev/null
    fi

    ensure_field "opportunity" "name" '{"api_name": "name", "label": "Opportunity Name", "type": "Text", "required": true, "is_name_field": true}'
    ensure_field "opportunity" "amount" '{"api_name": "amount", "label": "Amount", "type": "Currency"}'
    ensure_field "opportunity" "stage_name" '{"api_name": "stage_name", "label": "Stage", "type": "Picklist", "options": ["Prospecting", "Qualification", "Needs Analysis", "Proposal", "Negotiation", "Closed Won", "Closed Lost"]}'
    ensure_field "opportunity" "close_date" '{"api_name": "close_date", "label": "Close Date", "type": "Date"}'
    ensure_field "opportunity" "account_id" '{"api_name": "account_id", "label": "Account", "type": "Lookup", "reference_to": ["account"]}'

    api_patch "/api/metadata/objects/opportunity" '{"list_fields": ["name", "stage_name", "amount", "close_date", "account_id"]}' > /dev/null


    # ==========================
    # CASE
    # ==========================
    if ! api_get "/api/metadata/objects/case" | grep -q '"api_name":"case"'; then
        echo "   Creating Case object..."
        api_post "/api/metadata/objects" '{
            "label": "Case",
            "plural_label": "Cases",
            "api_name": "case",
            "description": "Case (Test Object)",
            "is_custom": true,
            "list_fields": ["subject"]
        }' > /dev/null
    fi

    ensure_field "case" "subject" '{"api_name": "subject", "label": "Subject", "type": "Text", "required": true, "is_name_field": true}'
    ensure_field "case" "status" '{"api_name": "status", "label": "Status", "type": "Picklist", "options": ["New", "Working", "Escalated", "Closed"]}'
    ensure_field "case" "priority" '{"api_name": "priority", "label": "Priority", "type": "Picklist", "options": ["Low", "Medium", "High", "Critical"]}'
    ensure_field "case" "contact_id" '{"api_name": "contact_id", "label": "Contact", "type": "Lookup", "reference_to": ["contact"]}'
    ensure_field "case" "account_id" '{"api_name": "account_id", "label": "Account", "type": "Lookup", "reference_to": ["account"]}'

    api_patch "/api/metadata/objects/case" '{"list_fields": ["subject", "status", "priority", "contact_id", "account_id"]}' > /dev/null


    # ==========================
    # APP
    # ==========================
    if ! api_get "/api/metadata/apps" | grep -q '"id":"nexus_crm"'; then
        echo "   Creating Nexus CRM App..."
        api_post "/api/metadata/apps" '{
            "id": "nexus_crm",
            "name": "nexus_crm",
            "label": "Nexus CRM",
            "description": "CRM",
            "icon": "LayoutDashboard",
            "color": "#2563eb",
            "navigation_items": [
                {"id": "nav_account", "type": "object", "object_api_name": "account", "label": "Accounts", "icon": "Users"},
                {"id": "nav_contact", "type": "object", "object_api_name": "contact", "label": "Contacts", "icon": "User"},
                {"id": "nav_lead", "type": "object", "object_api_name": "lead", "label": "Leads", "icon": "Filter"},
                {"id": "nav_opportunity", "type": "object", "object_api_name": "opportunity", "label": "Opportunities", "icon": "DollarSign"},
                {"id": "nav_case", "type": "object", "object_api_name": "case", "label": "Cases", "icon": "Briefcase"}
            ],
            "is_default": true
        }' > /dev/null
    fi

    echo "   ✅ Test objects verification completed."
}
