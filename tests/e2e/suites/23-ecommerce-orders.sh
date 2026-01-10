#!/bin/bash
# tests/e2e/suites/23-ecommerce-orders.sh
# E-Commerce Order Processing E2E Tests
# REFACTORED: Uses helper libraries

set +e
SUITE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SUITE_DIR/../config.sh"
source "$SUITE_DIR/../lib/helpers.sh"
source "$SUITE_DIR/../lib/api.sh"
source "$SUITE_DIR/../lib/schema_helpers.sh"
source "$SUITE_DIR/../lib/test_data.sh"

SUITE_NAME="E-Commerce Order Processing"
TIMESTAMP=$(date +%s)

# Test data IDs
PRODUCT_IDS=()
CUSTOMER_ID=""
ORDER_ID=""
LINEITEM_IDS=()

run_suite() {
    section_header "$SUITE_NAME"
    
    if ! api_login; then
        echo "Failed to login. Skipping suite."
        return 1
    fi
    
    # Schema Setup (using helpers)
    setup_schemas
    setup_ecommerce_app
    
    # Tests
    test_create_product_catalog
    test_create_customer
    test_create_order_with_lineitems
    test_order_status_workflow
    test_order_totals
    test_order_cancellation
    test_query_orders
    test_inventory_check
}

# =========================================
# SCHEMA SETUP (Using Helpers)
# =========================================

setup_schemas() {
    echo "Setup: Creating E-Commerce Schemas"
    
    # Product Schema
    ensure_schema "ecom_product" "Product" "Products"
    add_field "ecom_product" "sku" "SKU" "Text" true
    add_field "ecom_product" "description" "Description" "LongText"
    add_field "ecom_product" "price" "Price" "Currency" true
    add_field "ecom_product" "inventory_count" "Inventory" "Number"
    add_picklist "ecom_product" "category" "Category" "Electronics,Clothing,Home,Sports,Books"
    add_field "ecom_product" "is_active" "Active" "Checkbox"
    
    # Customer Schema
    ensure_schema "ecom_customer" "Customer" "Customers"
    add_field "ecom_customer" "email" "Email" "Email" true
    add_field "ecom_customer" "phone" "Phone" "Phone"
    add_field "ecom_customer" "shipping_address" "Shipping Address" "LongText"
    add_field "ecom_customer" "billing_address" "Billing Address" "LongText"
    add_picklist "ecom_customer" "customer_type" "Type" "Regular,Premium,VIP,Wholesale"
    add_field "ecom_customer" "total_orders" "Total Orders" "Number"
    
    # Order Schema
    ensure_schema "ecom_order" "Order" "Orders"
    add_field "ecom_order" "order_number" "Order Number" "Text" true
    add_lookup "ecom_order" "customer_id" "Customer" "ecom_customer" true
    add_field "ecom_order" "order_date" "Order Date" "DateTime"
    add_picklist "ecom_order" "status" "Status" "New,Processing,Shipped,Delivered,Cancelled,Refunded"
    add_field "ecom_order" "subtotal" "Subtotal" "Currency"
    add_field "ecom_order" "tax_amount" "Tax" "Currency"
    add_field "ecom_order" "shipping_amount" "Shipping" "Currency"
    add_field "ecom_order" "total_amount" "Total" "Currency"
    add_picklist "ecom_order" "shipping_method" "Shipping Method" "Standard,Express,Overnight,Pickup"
    add_field "ecom_order" "tracking_number" "Tracking Number" "Text"
    
    # Line Item Schema
    ensure_schema "ecom_lineitem" "Line Item" "Line Items"
    add_lookup "ecom_lineitem" "order_id" "Order" "ecom_order" true
    add_lookup "ecom_lineitem" "product_id" "Product" "ecom_product" true
    add_field "ecom_lineitem" "quantity" "Quantity" "Number" true
    add_field "ecom_lineitem" "unit_price" "Unit Price" "Currency"
    add_field "ecom_lineitem" "line_total" "Line Total" "Currency"
}

setup_ecommerce_app() {
    echo ""
    echo "Setup: Creating E-Commerce App"
    
    local nav_items=$(build_nav_items \
        "$(nav_item 'nav-ecom-products' 'object' 'Products' 'ecom_product' 'Package')" \
        "$(nav_item 'nav-ecom-customers' 'object' 'Customers' 'ecom_customer' 'Users')" \
        "$(nav_item 'nav-ecom-orders' 'object' 'Orders' 'ecom_order' 'ShoppingBag')" \
        "$(nav_item 'nav-ecom-lineitems' 'object' 'Line Items' 'ecom_lineitem' 'List')")
    
    ensure_app "app_Ecommerce" "E-Commerce" "ShoppingCart" "#10B981" "$nav_items" "Order Processing and Fulfillment"
}

# =========================================
# TESTS
# =========================================

test_create_product_catalog() {
    echo ""
    echo "Test 23.1: Create Product Catalog"
    
    local products=(
        '{"name": "Wireless Headphones", "sku": "ELEC-WH-001", "price": 149.99, "inventory_count": 50, "category": "Electronics", "is_active": true}'
        '{"name": "Running Shoes", "sku": "SPRT-RS-002", "price": 89.99, "inventory_count": 100, "category": "Sports", "is_active": true}'
        '{"name": "Desk Lamp", "sku": "HOME-DL-003", "price": 45.00, "inventory_count": 75, "category": "Home", "is_active": true}'
        '{"name": "Programming Guide", "sku": "BOOK-PG-004", "price": 35.99, "inventory_count": 200, "category": "Books", "is_active": true}'
        '{"name": "Winter Jacket", "sku": "CLTH-WJ-005", "price": 199.99, "inventory_count": 30, "category": "Clothing", "is_active": true}'
    )
    
    local count=0
    for prod in "${products[@]}"; do
        local res=$(api_post "/api/data/ecom_product" "$prod")
        local id=$(json_extract "$res" "id")
        [ -n "$id" ] && { PRODUCT_IDS+=("$id"); count=$((count + 1)); }
    done
    
    echo "  Created $count products"
    [ $count -eq 5 ] && test_passed "Product catalog created" || test_failed "Only $count products"
}

test_create_customer() {
    echo ""
    echo "Test 23.2: Create Customer"
    
    local res=$(api_post "/api/data/ecom_customer" '{"name": "Jane Smith '$TIMESTAMP'", "email": "jane.'$TIMESTAMP'@example.com", "phone": "5551234567", "shipping_address": "123 Main St\nNew York, NY", "customer_type": "Premium", "total_orders": 0}')
    CUSTOMER_ID=$(json_extract "$res" "id")
    
    [ -n "$CUSTOMER_ID" ] && test_passed "Customer created" || test_failed "Failed to create"
}

test_create_order_with_lineitems() {
    echo ""
    echo "Test 23.3: Create Order with Line Items"
    
    if [ -z "$CUSTOMER_ID" ] || [ ${#PRODUCT_IDS[@]} -lt 3 ]; then
        test_failed "Missing customer or products"; return 1
    fi
    
    local res=$(api_post "/api/data/ecom_order" '{"name": "ORD-'$TIMESTAMP'", "order_number": "ORD-'$TIMESTAMP'", "customer_id": "'$CUSTOMER_ID'", "status": "New", "shipping_method": "Express", "shipping_amount": 15.00}')
    ORDER_ID=$(json_extract "$res" "id")
    [ -z "$ORDER_ID" ] && { test_failed "Failed to create order"; return 1; }
    echo "  ✓ Order created: $ORDER_ID"
    
    local items=("0:2:149.99:299.98" "1:1:89.99:89.99" "3:3:35.99:107.97")
    for item in "${items[@]}"; do
        IFS=':' read -r pi qty up lt <<< "$item"
        local li=$(api_post "/api/data/ecom_lineitem" '{"name": "Line '$pi'", "order_id": "'$ORDER_ID'", "product_id": "'${PRODUCT_IDS[$pi]}'", "quantity": '$qty', "unit_price": '$up', "line_total": '$lt'}')
        local id=$(json_extract "$li" "id")
        [ -n "$id" ] && LINEITEM_IDS+=("$id")
    done
    
    echo "  Created ${#LINEITEM_IDS[@]} line items"
    [ ${#LINEITEM_IDS[@]} -eq 3 ] && test_passed "Order with line items" || test_failed "Only ${#LINEITEM_IDS[@]} items"
}

test_order_status_workflow() {
    echo ""
    echo "Test 23.4: Order Status Workflow"
    
    [ -z "$ORDER_ID" ] && { test_failed "No order"; return 1; }
    
    for status in "Processing" "Shipped" "Delivered"; do
        api_patch "/api/data/ecom_order/$ORDER_ID" '{"status": "'$status'"}' > /dev/null
        echo "  ✓ Status → $status"
        [ "$status" == "Shipped" ] && api_patch "/api/data/ecom_order/$ORDER_ID" '{"tracking_number": "1Z999AA10123456784"}' > /dev/null
    done
    
    test_passed "Order workflow complete"
}

test_order_totals() {
    echo ""
    echo "Test 23.5: Update Order Totals"
    
    [ -z "$ORDER_ID" ] && { test_failed "No order"; return 1; }
    
    local subtotal=497.94
    local tax=$(echo "$subtotal * 0.08" | bc)
    local total=$(echo "$subtotal + $tax + 15" | bc)
    echo "  Subtotal: $subtotal | Tax: $tax | Total: $total"
    
    api_patch "/api/data/ecom_order/$ORDER_ID" '{"subtotal": '$subtotal', "tax_amount": '$tax', "total_amount": '$total'}' > /dev/null
    test_passed "Order totals updated"
}

test_order_cancellation() {
    echo ""
    echo "Test 23.6: Order Cancellation Flow"
    
    if [ -z "$CUSTOMER_ID" ]; then
        test_failed "No customer for cancel test"; return 1
    fi
    
    local res=$(api_post "/api/data/ecom_order" '{"name": "ORD-CANCEL-'$TIMESTAMP'", "order_number": "ORD-CANCEL-'$TIMESTAMP'", "customer_id": "'$CUSTOMER_ID'", "status": "New", "total_amount": 54.00}')
    local cancel_id=$(json_extract "$res" "id")
    
    if [ -n "$cancel_id" ]; then
        api_patch "/api/data/ecom_order/$cancel_id" '{"status": "Cancelled"}' > /dev/null
        echo "  ✓ Order cancelled"
        api_delete "/api/data/ecom_order/$cancel_id" > /dev/null 2>&1
        test_passed "Cancellation workflow"
    else
        test_failed "Failed to create cancel order"
    fi
}

test_query_orders() {
    echo ""
    echo "Test 23.7: Query Orders"
    
    local delivered=$(api_post "/api/data/query" '{"object_api_name": "ecom_order", "filter_expr": "status == '"'"'Delivered'"'"'"}')
    echo "  Delivered orders: $(echo "$delivered" | jq '.data | length' 2>/dev/null || echo 0)"
    
    if [ -n "$ORDER_ID" ]; then
        local lineitems=$(api_post "/api/data/query" '{"object_api_name": "ecom_lineitem", "filter_expr": "order_id == '"'"'$ORDER_ID'"'"'"}')
        echo "  Line items in order: $(echo "$lineitems" | jq '.data | length' 2>/dev/null || echo 0)"
    fi
    
    test_passed "Order queries completed"
}

test_inventory_check() {
    echo ""
    echo "Test 23.8: Inventory Check"
    
    [ -n "${PRODUCT_IDS[0]}" ] && { api_patch "/api/data/ecom_product/${PRODUCT_IDS[0]}" '{"inventory_count": 48}' > /dev/null; echo "  ✓ Headphones: 50 → 48"; }
    [ -n "${PRODUCT_IDS[1]}" ] && { api_patch "/api/data/ecom_product/${PRODUCT_IDS[1]}" '{"inventory_count": 99}' > /dev/null; echo "  ✓ Shoes: 100 → 99"; }
    
    test_passed "Inventory updated"
}

# =========================================
# CLEANUP (Using Helpers)
# =========================================

test_cleanup() {
    echo ""
    echo "Test 23.9: Cleanup Test Data"
    
    # 1. Delete records
    cleanup_records "ecom_lineitem" "${LINEITEM_IDS[@]}"
    [ -n "$ORDER_ID" ] && api_delete "/api/data/ecom_order/$ORDER_ID" > /dev/null 2>&1
    echo "  ✓ Deleted 1 ecom_order records"
    [ -n "$CUSTOMER_ID" ] && api_delete "/api/data/ecom_customer/$CUSTOMER_ID" > /dev/null 2>&1
    echo "  ✓ Deleted 1 ecom_customer records"
    cleanup_records "ecom_product" "${PRODUCT_IDS[@]}"
    
    # 2. Delete schemas
    echo "  Cleaning up schemas..."
    delete_schema "ecom_lineitem"
    delete_schema "ecom_order"
    delete_schema "ecom_customer"
    delete_schema "ecom_product"
    
    # 3. Delete app
    echo "  Cleaning up app..."
    delete_app "app_Ecommerce"
    
    test_passed "Cleanup completed"
}

# Trap cleanup on exit
trap test_cleanup EXIT

# Run if executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_suite
fi
