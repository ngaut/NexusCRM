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

def create_app(token, app_id, label, color):
    data = {"id": app_id, "label": label, "color": color, "icon": "briefcase"}
    req = urllib.request.Request(f"{BASE_URL}/metadata/apps", data=json.dumps(data).encode(), 
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    try:
        urllib.request.urlopen(req)
        print(f"✅ App '{label}' created.")
    except urllib.error.HTTPError as e:
        if e.code == 409 or e.code == 500: 
            print(f"ℹ️ App '{label}' likely already exists (Code {e.code}). Proceeding to verify...")
        else: raise e

    # Verify ID and Color via GET
    req = urllib.request.Request(f"{BASE_URL}/metadata/apps", headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        apps = json.loads(resp.read())["apps"]
        found = next((a for a in apps if a["id"] == app_id), None)

        if not found: raise Exception(f"App {app_id} not found after creation attempt!")
        if found.get("color") != color: 
             print(f"⚠️ Warning: App color mismatch. Expected {color}, got {found.get('color')}")
        else:print(f"✅ App '{label}' verified with color {color}.")

def create_record(token, object_name, payload):
    req = urllib.request.Request(f"{BASE_URL}/data/{object_name}", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        body = json.loads(resp.read())
        # print(f"DEBUG Response: {body}")
        if "data" in body:
            return body["data"]
        if "record" in body:
            return body["record"]
        return body

def main():
    try:
        print("🔑 Logging in...")
        token = login()
        
        print("📱 Creating 'Nexus Sales' App...")
        create_app(token, "nexus_sales", "Nexus Sales", "blue")

        print("🏢 Creating Account 'Acme Corp'...")
        account = create_record(token, "account", {"name": "Acme Corp", "industry": "Technology"})
        print(f"✅ Account Created: {account['id']}")

        print("busts Creating Contact 'Alice Smith'...")
        contact = create_record(token, "contact", {
            "name": "Alice Smith", 
            "email": "alice@acme.com", 
            "account_id": account["id"]
        })
        print(f"✅ Contact Created: {contact['id']} (Linked to Account {contact.get('account_id')})")

        if contact.get('account_id') == account['id']:
            print("SUCCESS: Phase 1 (CRM Setup) Verified.")
        else:
            print("FAILURE: Contact not linked to Account.")
            sys.exit(1)

    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
