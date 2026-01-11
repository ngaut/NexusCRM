
import sys
import os
import json
import uuid

# Add parent directory to path to import migration_tool modules
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from backend.scripts.migration_tool.client import NexusClient

HOST = "localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

APP_ID = "data_platform"

NAV_ITEMS = [
    {"type": "dashboard", "target": "sales_dashboard", "label": "Sales Dashboard", "icon": "Layout"},
    {"type": "object", "target": "lead_history", "label": "Lead History (2.7M)", "icon": "History"},
    {"type": "object", "target": "campaign_member", "label": "Campaign Members", "icon": "Users"},
    {"type": "object", "target": "task", "label": "Tasks", "icon": "CheckSquare"},
    {"type": "object", "target": "fivetran_audit", "label": "Fivetran Audit", "icon": "FileText"},
    {"type": "object", "target": "lead", "label": "Leads", "icon": "UserPlus"},
    {"type": "object", "target": "qingflow", "label": "Qingflow", "icon": "Activity"},
    {"type": "object", "target": "salesforce_task", "label": "SF Tasks", "icon": "Clipboard"},
    {"type": "object", "target": "account", "label": "Accounts", "icon": "Briefcase"},
    {"type": "object", "target": "opportunity", "label": "Opportunities", "icon": "DollarSign"},
    {"type": "object", "target": "contact", "label": "Contacts", "icon": "User"},
    {"type": "object", "target": "org_history", "label": "Org History", "icon": "Clock"},
    {"type": "object", "target": "global_arr_forecast_c", "label": "ARR Forecast", "icon": "TrendingUp"}
]

def login():
    client = NexusClient(f"http://{HOST}", "")
    data, err = client.request("POST", "/api/auth/login", {"email": EMAIL, "password": PASSWORD})
    if err:
        print(f"Login failed: {err}")
        return None
    return data["token"]

def configure_app():
    token = login()
    if not token: return
    
    client = NexusClient(f"http://{HOST}", token)
    
    # 1. Fetch ALL apps
    print(f"Fetching app list...")
    resp, err = client.request("GET", "/api/metadata/apps")
    if err:
        print(f"Error fetching apps: {err}")
        return
    
    apps = resp.get("data", [])
    target_app = None
    for app in apps:
        if app.get("id") == APP_ID:
            target_app = app
            break
            
    if not target_app:
        print(f"❌ App '{APP_ID}' not found in the list.")
        return

    print(f"Found App: {target_app['label']} (ID: {target_app['id']})")
    
    # 2. Construct Navigation Items
    new_nav_items = []
    
    for item in NAV_ITEMS:
        nav_entry = {
            "id": str(uuid.uuid4()),
            "type": item["type"],
            "label": item["label"],
            "icon": item["icon"]
        }
        
        if item["type"] == "object":
            nav_entry["object_api_name"] = item["target"]
        elif item["type"] == "dashboard":
            nav_entry["dashboard_id"] = item["target"]
            
        new_nav_items.append(nav_entry)
        
    # 3. Patch the App
    payload = {
        "navigation_items": new_nav_items
    }
    
    print(f"Updating app {APP_ID} with {len(new_nav_items)} items...")
    
    # API likely accepts partial update for PATCH
    update_resp, err = client.request("PATCH", f"/api/metadata/apps/{APP_ID}", payload)
    
    if err:
        print(f"❌ Update failed: {err}")
    else:
        print("✅ App configuration updated successfully!")
        # print(json.dumps(update_resp, indent=2))

if __name__ == "__main__":
    configure_app()
