import urllib.request
import json
import sys

BASE_URL = "http://localhost:3001/api"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

def login():
    data = json.dumps({"email": EMAIL, "password": PASSWORD}).encode()
    req = urllib.request.Request(f"{BASE_URL}/auth/login", data=data, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read())["token"]

def update_app(token, app_id):
    items = [
        {"id": "nav_acc", "type": "object", "object_api_name": "account", "label": "Accounts", "icon": "users"},
        {"id": "nav_con", "type": "object", "object_api_name": "contact", "label": "Contacts", "icon": "user"},
        {"id": "nav_tick", "type": "object", "object_api_name": "ticket", "label": "Tickets", "icon": "ticket"}
    ]
    payload = {"navigation_items": items}
    
    # Use PUT or PATCH? Usually PATCH for partial updates, but AppConfig might need full replacement of list.
    # handler_metadata_ui.go UpdateApp uses PATCH.
    
    req = urllib.request.Request(f"{BASE_URL}/metadata/apps/{app_id}", data=json.dumps(payload).encode(), method="PATCH",
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    try:
        urllib.request.urlopen(req)
        print(f"✅ Updated app {app_id} navigation items.")
    except urllib.error.HTTPError as e:
        print(f"❌ Failed to update app {app_id}: {e}")
        print(e.read())
        sys.exit(1)

def main():
    token = login()
    update_app(token, "nexus_sales")

if __name__ == "__main__":
    main()
