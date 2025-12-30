#!/bin/bash
# tests/e2e/lib/setup.sh
# Ensures standard business objects exist before running tests

source "$(dirname "${BASH_SOURCE[0]}")/api.sh"

ensure_standard_objects_exist() {
    echo "⚡️ Ensuring standard business objects exist..."

    # ACCOUNT
    if ! api_get "/api/metadata/schemas/account" | grep -q '"api_name":"account"'; then
        echo "   Creating Account object..."
        api_post "/api/metadata/schemas" '{
            "label": "Account",
            "plural_label": "Accounts",
            "api_name": "account",
            "description": "Standard Account Object",
            "is_custom": false,
            "list_fields": ["name", "industry", "type", "website", "phone"]
        }' > /dev/null
        
        # Core Fields
        api_post "/api/metadata/schemas/account/fields" '{"api_name": "name", "label": "Account Name", "type": "Text", "required": true, "is_name_field": true}' > /dev/null
        api_post "/api/metadata/schemas/account/fields" '{"api_name": "industry", "label": "Industry", "type": "Select", "options": ["Technology", "Finance", "Healthcare"]}' > /dev/null
        api_post "/api/metadata/schemas/account/fields" '{"api_name": "type", "label": "Type", "type": "Select", "options": ["Customer", "Partner", "Prospect"]}' > /dev/null
        api_post "/api/metadata/schemas/account/fields" '{"api_name": "website", "label": "Website", "type": "URL"}' > /dev/null
        api_post "/api/metadata/schemas/account/fields" '{"api_name": "annual_revenue", "label": "Annual Revenue", "type": "Currency"}' > /dev/null
        api_post "/api/metadata/schemas/account/fields" '{"api_name": "phone", "label": "Phone", "type": "Phone"}' > /dev/null
        api_post "/api/metadata/schemas/account/fields" '{"api_name": "theme_color", "label": "Theme Color", "type": "Text"}' > /dev/null
    fi

    # CONTACT
    if ! api_get "/api/metadata/schemas/contact" | grep -q '"api_name":"contact"'; then
        echo "   Creating Contact object..."
        api_post "/api/metadata/schemas" '{
            "label": "Contact",
            "plural_label": "Contacts",
            "api_name": "contact",
            "description": "Standard Contact Object",
            "is_custom": false,
            "list_fields": ["first_name", "last_name", "email", "phone", "title", "account_id"]
        }' > /dev/null
        
        api_post "/api/metadata/schemas/contact/fields" '{"api_name": "first_name", "label": "First Name", "type": "Text"}' > /dev/null
        api_post "/api/metadata/schemas/contact/fields" '{"api_name": "last_name", "label": "Last Name", "type": "Text", "required": true, "is_name_field": true}' > /dev/null
        api_post "/api/metadata/schemas/contact/fields" '{"api_name": "email", "label": "Email", "type": "Email"}' > /dev/null
        api_post "/api/metadata/schemas/contact/fields" '{"api_name": "phone", "label": "Phone", "type": "Phone"}' > /dev/null
        api_post "/api/metadata/schemas/contact/fields" '{"api_name": "title", "label": "Title", "type": "Text"}' > /dev/null
        api_post "/api/metadata/schemas/contact/fields" '{"api_name": "account_id", "label": "Account", "type": "Lookup", "reference_to": ["account"]}' > /dev/null
    fi

    # LEAD
    if ! api_get "/api/metadata/schemas/lead" | grep -q '"api_name":"lead"'; then
        echo "   Creating Lead object..."
        api_post "/api/metadata/schemas" '{
            "label": "Lead",
            "plural_label": "Leads",
            "api_name": "lead",
            "description": "Standard Lead Object",
            "is_custom": false,
            "list_fields": ["name", "company", "status", "email"]
        }' > /dev/null
        
        api_post "/api/metadata/schemas/lead/fields" '{"api_name": "company", "label": "Company", "type": "Text", "required": true}' > /dev/null
        api_post "/api/metadata/schemas/lead/fields" '{"api_name": "name", "label": "Name", "type": "Text", "required": true, "is_name_field": true}' > /dev/null
        api_post "/api/metadata/schemas/lead/fields" '{"api_name": "email", "label": "Email", "type": "Email"}' > /dev/null
        api_post "/api/metadata/schemas/lead/fields" '{"api_name": "status", "label": "Status", "type": "Select", "options": ["Open", "Contacted", "Qualified", "Unqualified"]}' > /dev/null
    fi

    # OPPORTUNITY
    if ! api_get "/api/metadata/schemas/opportunity" | grep -q '"api_name":"opportunity"'; then
        echo "   Creating Opportunity object..."
        api_post "/api/metadata/schemas" '{
            "label": "Opportunity",
            "plural_label": "Opportunities",
            "api_name": "opportunity",
            "description": "Standard Opportunity Object",
            "is_custom": false,
            "list_fields": ["name", "stage_name", "amount", "close_date", "account_id"]
        }' > /dev/null
        
        api_post "/api/metadata/schemas/opportunity/fields" '{"api_name": "name", "label": "Opportunity Name", "type": "Text", "required": true, "is_name_field": true}' > /dev/null
        api_post "/api/metadata/schemas/opportunity/fields" '{"api_name": "amount", "label": "Amount", "type": "Currency"}' > /dev/null
        api_post "/api/metadata/schemas/opportunity/fields" '{"api_name": "stage_name", "label": "Stage", "type": "Select", "options": ["Prospecting", "Qualification", "Needs Analysis", "Proposal", "Negotiation", "Closed Won", "Closed Lost"]}' > /dev/null
        api_post "/api/metadata/schemas/opportunity/fields" '{"api_name": "close_date", "label": "Close Date", "type": "Date"}' > /dev/null
        api_post "/api/metadata/schemas/opportunity/fields" '{"api_name": "account_id", "label": "Account", "type": "Lookup", "reference_to": ["account"]}' > /dev/null
    fi

    # CASE
    if ! api_get "/api/metadata/schemas/case" | grep -q '"api_name":"case"'; then
        echo "   Creating Case object..."
        api_post "/api/metadata/schemas" '{
            "label": "Case",
            "plural_label": "Cases",
            "api_name": "case",
            "description": "Standard Service Case",
            "is_custom": false,
            "list_fields": ["subject", "status", "priority", "contact_id", "account_id"]
        }' > /dev/null

        api_post "/api/metadata/schemas/case/fields" '{"api_name": "subject", "label": "Subject", "type": "Text", "required": true, "is_name_field": true}' > /dev/null
        api_post "/api/metadata/schemas/case/fields" '{"api_name": "status", "label": "Status", "type": "Select", "options": ["New", "Working", "Escalated", "Closed"]}' > /dev/null
        api_post "/api/metadata/schemas/case/fields" '{"api_name": "contact_id", "label": "Contact", "type": "Lookup", "reference_to": ["contact"]}' > /dev/null
        api_post "/api/metadata/schemas/case/fields" '{"api_name": "account_id", "label": "Account", "type": "Lookup", "reference_to": ["account"]}' > /dev/null
        api_post "/api/metadata/schemas/case/fields" '{"api_name": "priority", "label": "Priority", "type": "Select", "options": ["Low", "Medium", "High", "Critical"]}' > /dev/null
    fi

    echo "   ✅ Standard objects verification completed."
}
