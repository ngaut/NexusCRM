import urllib.request
import json
import sys
import time

BASE_URL = "http://localhost:3001/api"
# Credentials from setup_ticket_system.py
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

def login():
    print("🔑 Logging in...")
    data = json.dumps({"email": EMAIL, "password": PASSWORD}).encode()
    req = urllib.request.Request(f"{BASE_URL}/auth/login", data=data, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req) as resp:
            return json.loads(resp.read())["token"]
    except Exception as e:
        print(f"❌ Login failed: {e}")
        sys.exit(1)

def create_app(token):
    print("📱 Creating CRM App...")
    payload = {
        "id": "crm",
        "label": "Nexus CRM",
        "description": "Core CRM Application",
        "icon": "Box",
        "color": "indigo",
        "is_default": True,
        "navigation_items": [
            {"type": "object", "object_api_name": "account", "label": "Accounts", "icon": "Building"},
            {"type": "object", "object_api_name": "contact", "label": "Contacts", "icon": "User"},
            {"type": "object", "object_api_name": "lead", "label": "Leads", "icon": "Filter"},
            {"type": "object", "object_api_name": "opportunity", "label": "Opportunities", "icon": "DollarSign"}
        ]
    }
    
    req = urllib.request.Request(
        f"{BASE_URL}/metadata/apps", 
        data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"}
    )
    
    try:
        urllib.request.urlopen(req)
        print("✅ App 'Nexus CRM' created successfully")
    except urllib.error.HTTPError as e:
        if e.code == 409:
            print("ℹ️ App 'crm' already exists")
        else:
            print(f"❌ Failed to create app: {e} - {e.read().decode()}")
            sys.exit(1)

def main():
    # Wait for server to be ready
    print("⏳ Waiting for server API...")
    for i in range(30):
        try:
            urllib.request.urlopen(f"http://localhost:3001/health")
            break
        except:
            time.sleep(1)
            
    token = login()
    create_app(token)

if __name__ == "__main__":
    main()
