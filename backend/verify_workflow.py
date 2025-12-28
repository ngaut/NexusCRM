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
    payload = {"object_api_name": "account", "criteria": [{"field": "name", "op": "=", "val": name}]}
    req = urllib.request.Request(f"{BASE_URL}/data/query", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        body = json.loads(resp.read())
        records = body.get("records", [])
        if records: return records[0]["id"]
    return None

def create_flow(token):
    # Flow: If Priority is High, Set Status to In Progress
    payload = {
        "name": "Auto-Prioritize Ticket",
        "status": "Active",
        "trigger_object": "ticket",
        "trigger_type": "beforeCreate",
        # Formula: priority == 'High'. Using double equals for safety if govaluate based.
        "trigger_condition": "priority == \"High\"", 
        "action_type": "updateRecord",
        "action_config": {
            "updates": {
                "status": "In Progress"
            }
        }
    }
    
    req = urllib.request.Request(f"{BASE_URL}/metadata/flows", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    try:
        urllib.request.urlopen(req)
        print("✅ Created Workflow: Auto-Prioritize Ticket")
    except urllib.error.HTTPError as e:
        if e.code == 500 or e.code == 409:
             print(f"ℹ️ Flow likely already exists (Code {e.code})")
        else:
             print(f"❌ Failed to create flow: {e}")

def create_test_ticket(token, acc_id):
    payload = {
        "name": "High Priority Test",
        "priority": "High",
        "status": "New", # Should change to In Progress
        "description": "Testing Workflow",
        "customer_id": acc_id
    }
    req = urllib.request.Request(f"{BASE_URL}/data/ticket", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req) as resp:
        body = json.loads(resp.read())
        record = body.get("record", body.get("data"))
        return record

def main():
    token = login()
    acc_id = get_account_id(token, "Acme Corp")
    if not acc_id:
        print("Acmecorp not found")
        sys.exit(1)

    create_flow(token)
    
    print("Creating Ticket with Priority=High, Status=New...")
    ticket = create_test_ticket(token, acc_id)
    
    print(f"Ticket Created: {ticket['status']}")
    
    if ticket['status'] == 'In Progress':
        print("SUCCESS: Workflow updated status to In Progress")
    else:
        print(f"FAILURE: Status is {ticket['status']}, expected In Progress")
        sys.exit(1)

if __name__ == "__main__":
    main()
