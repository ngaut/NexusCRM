#!/bin/bash
# tests/e2e/lib/schema_helpers.sh
# Reusable schema and field creation helpers to reduce code duplication

# Get directory of this file
SCHEMA_HELPERS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source dependencies if not loaded
if [ -z "$BASE_URL" ]; then
    source "$SCHEMA_HELPERS_DIR/../config.sh"
fi
if ! type api_post &>/dev/null; then
    source "$SCHEMA_HELPERS_DIR/api.sh"
fi

# =========================================
# SCHEMA HELPERS
# =========================================

# Create schema if it doesn't exist
# Usage: ensure_schema "api_name" "label" "plural_label"
ensure_schema() {
    local api_name="$1"
    local label="$2"
    local plural_label="${3:-${label}s}"
    
    local check=$(api_get "/api/metadata/objects/$api_name")
    
    # API returns {"schema": {...}} for existing schemas
    if echo "$check" | jq -e '.schema.api_name' &>/dev/null; then
        echo "  ✓ $label schema already exists"
        return 0
    fi
    
    local payload=$(jq -n \
        --arg label "$label" \
        --arg api_name "$api_name" \
        --arg plural "$plural_label" \
        '{label: $label, api_name: $api_name, plural_label: $plural, is_custom: true}')
    
    local res=$(api_post "/api/metadata/objects" "$payload")
    
    if echo "$res" | jq -e '.api_name // .schema.api_name' &>/dev/null; then
        echo "  ✓ $label schema created"
        return 0
    else
        local err=$(echo "$res" | jq -r '.error // empty')
        if [ -z "$err" ]; then err="$res"; fi
        echo "  ✗ Failed to create $label schema: $err"
        return 1
    fi
}

# Delete schema
# Usage: delete_schema "api_name"
delete_schema() {
    local api_name="$1"
    
    local res=$(api_delete "/api/metadata/objects/$api_name")
    
    # Check if delete was successful (200 OK or 404 Not Found is cleaner than error)
    if [ -z "$res" ] || echo "$res" | grep -q "\"success\":true" || echo "$res" | grep -q "deleted"; then
         echo "  ✓ Deleted schema $api_name"
         return 0
    elif echo "$res" | grep -q "not found"; then
         echo "  - Schema $api_name not found (already deleted)"
         return 0
    else
         echo "  Warning: Failed to delete schema $api_name: $res"
         return 1
    fi
}

# Add field to schema
# Usage: add_field "schema_api_name" "field_api_name" "label" "type" [required] [extra_json]
# Example: add_field "hr_employee" "email" "Email" "Email" true
# Example: add_field "hr_employee" "dept" "Department" "Lookup" false '{"reference_to":["hr_dept"]}'
add_field() {
    local schema="$1"
    local api_name="$2"
    local label="$3"
    local type="$4"
    local required="${5:-false}"
    local extra="${6:-}"
    
    # Build base JSON
    local json=$(jq -n \
        --arg api_name "$api_name" \
        --arg label "$label" \
        --arg type "$type" \
        --argjson required "$required" \
        '{api_name: $api_name, label: $label, type: $type, required: $required}')
    
    # Merge extra JSON if provided
    if [ -n "$extra" ]; then
        json=$(echo "$json" | jq --argjson extra "$extra" '. + $extra')
    fi
    
    local res=$(api_post "/api/metadata/objects/$schema/fields" "$json")
    
    if echo "$res" | jq -e '.api_name // .field.api_name' &>/dev/null; then
        return 0
    else
        local err=$(echo "$res" | jq -r '.error // empty')
        if [ -z "$err" ]; then err="$res"; fi
        # Don't warn about already exists
        if [[ "$err" != *"already exists"* ]]; then
            echo "  Warning: Failed to add field $api_name: $err"
        fi
        return 1
    fi
}

# Add picklist field
# Usage: add_picklist "schema" "api_name" "label" "Option1,Option2,Option3" [required]
add_picklist() {
    local schema="$1"
    local api_name="$2"
    local label="$3"
    local options_csv="$4"
    local required="${5:-false}"
    
    # Convert CSV to JSON array
    local options_json=$(echo "$options_csv" | tr ',' '\n' | jq -R . | jq -s .)
    
    add_field "$schema" "$api_name" "$label" "Picklist" "$required" "{\"options\": $options_json}"
}

# Add lookup field
# Usage: add_lookup "schema" "api_name" "label" "target_object" [required]
add_lookup() {
    local schema="$1"
    local api_name="$2"
    local label="$3"
    local target="$4"
    local required="${5:-false}"
    
    add_field "$schema" "$api_name" "$label" "Lookup" "$required" "{\"reference_to\": [\"$target\"]}"
}

# =========================================
# APP HELPERS
# =========================================

# Create app if it doesn't exist
# Usage: ensure_app "app_id" "label" "icon" "color" "nav_items_json"
ensure_app() {
    local app_id="$1"
    local label="$2"
    local icon="$3"
    local color="$4"
    local nav_items="$5"
    local description="${6:-$label Management}"
    
    local apps=$(api_get "/api/metadata/apps")
    
    if echo "$apps" | jq -e ".apps[]? | select(.id == \"$app_id\")" &>/dev/null; then
        echo "  ✓ $label App already exists"
        return 0
    fi
    
    local payload=$(jq -n \
        --arg id "$app_id" \
        --arg label "$label" \
        --arg desc "$description" \
        --arg icon "$icon" \
        --arg color "$color" \
        --argjson nav "$nav_items" \
        '{id: $id, label: $label, description: $desc, icon: $icon, color: $color, navigation_items: $nav, is_default: false}')
    
    local res=$(api_post "/api/metadata/apps" "$payload")
    
    if echo "$res" | jq -e '.id // .app.id' &>/dev/null; then
        echo "  ✓ $label App created"
        return 0
    else
        local err=$(echo "$res" | jq -r '.error // empty')
        if [ -z "$err" ]; then err="$res"; fi
        echo "  ✗ Failed to create $label App: $err"
        return 1
    fi
}

# Delete app
# Usage: delete_app "app_id"
delete_app() {
    local app_id="$1"
    
    local res=$(api_delete "/api/metadata/apps/$app_id")
    
    if [ -z "$res" ] || echo "$res" | grep -q "\"success\":true" || echo "$res" | grep -q "deleted"; then
         echo "  ✓ Deleted app $app_id"
         return 0
    elif echo "$res" | grep -q "not found"; then
         echo "  - App $app_id not found (already deleted)"
         return 0
    else
         echo "  Warning: Failed to delete app $app_id: $res"
         return 1
    fi
}

# Build navigation item JSON
# Usage: nav_item "id" "type" "label" "object_api_name" "icon"
nav_item() {
    local id="$1"
    local type="$2"
    local label="$3"
    local object="$4"
    local icon="$5"
    
    jq -n \
        --arg id "$id" \
        --arg type "$type" \
        --arg label "$label" \
        --arg object "$object" \
        --arg icon "$icon" \
        '{id: $id, type: $type, label: $label, object_api_name: $object, icon: $icon}'
}

# Build navigation items array from multiple nav_item calls
# Usage: nav_items=$(build_nav_items \
#   "$(nav_item 'id1' 'object' 'Label1' 'obj1' 'Icon1')" \
#   "$(nav_item 'id2' 'object' 'Label2' 'obj2' 'Icon2')")
build_nav_items() {
    local items=("$@")
    local json="["
    local first=true
    
    for item in "${items[@]}"; do
        if [ "$first" = true ]; then
            first=false
        else
            json+=","
        fi
        json+="$item"
    done
    json+="]"
    
    echo "$json"
}
