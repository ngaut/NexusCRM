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

def get_account_id(token, name):
    # Retrieve account by QUERY
    payload = {"object_api_name": "account", "criteria": [{"field": "name", "op": "=", "val": name}]}
    req = urllib.request.Request(f"{BASE_URL}/data/query", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        body = json.loads(resp.read())
        records = body.get("records", [])
        if records:
            return records[0]["id"]
    return None

def create_ticket(token, account_id):
    payload = {
        "name": "API Ticket",  # Name field is standard? Yes, `is_name_field` usually `name`.
        # Wait, Custom Objects have a Name field? Default is Name (Text).
        # My setup script didn't explicitly create 'name', but `_System_Object` usually ensures a name field?
        # `CreateSchema` adds default Name field if not present?
        # Re-checking standard fields.
        "priority": "High",
        "status": "New",
        "description": "Created via API",
        "customer_id": account_id
    }
    req = urllib.request.Request(f"{BASE_URL}/data/ticket", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    try:
        with urllib.request.urlopen(req) as resp:
            body = json.loads(resp.read())
            print(f"✅ Ticket Created: {body}")
            return body.get("record", body.get("data"))
    except urllib.error.HTTPError as e:
        print(f"❌ Failed to create ticket: {e}")
        print(e.read())
        sys.exit(1)

def main():
    token = login()
    acc_id = get_account_id(token, "Acme Corp")
    if not acc_id:
        print("❌ Acme Corp account not found (needed for linking)")
        sys.exit(1)
        
    print(f"Found Account ID: {acc_id}")
    create_ticket(token, acc_id)

if __name__ == "__main__":
    main()
