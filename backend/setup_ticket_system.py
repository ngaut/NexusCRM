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

def delete_object(token, api_name):
    req = urllib.request.Request(f"{BASE_URL}/metadata/objects/{api_name}", method="DELETE", 
        headers={"Authorization": f"Bearer {token}"})
    try:
        urllib.request.urlopen(req)
        print(f"xx Deleted object {api_name}")
    except urllib.error.HTTPError as e:
        print(f"ℹ️ Could not delete {api_name} (Code {e.code})")

def create_object(token, label, plural, api_name):
    payload = {
        "label": label,
        "plural_label": plural,
        "api_name": api_name,
        "description": f"Custom object: {label}",
        "is_custom": True
    }
    req = urllib.request.Request(f"{BASE_URL}/metadata/objects", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    try:
        urllib.request.urlopen(req)
        print(f"✅ Created object {api_name}")
    except urllib.error.HTTPError as e:
        if e.code == 409 or e.code == 500: print(f"ℹ️ Object {api_name} likely already exists")
        else: raise e

def create_field(token, obj_name, field_def):
    req = urllib.request.Request(f"{BASE_URL}/metadata/objects/{obj_name}/fields", data=json.dumps(field_def).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    try:
        urllib.request.urlopen(req)
        print(f"✅ Created field {obj_name}.{field_def['api_name']}")
    except urllib.error.HTTPError as e:
        if e.code == 409 or e.code == 500: print(f"ℹ️ Field {field_def['api_name']} likely already exists")
        else: 
            print(f"❌ Failed to create field {field_def['api_name']}: {e}")
            # continue instead of raise to try next fields

def create_list_view(token, obj_name):
    payload = {
        "object_api_name": obj_name,
        "label": "All " + obj_name.capitalize() + "s",
        "fields": ["name", "priority", "status", "created_date", "owner_id"],
        "filters": []
    }
    req = urllib.request.Request(f"{BASE_URL}/metadata/listviews", data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
    try:
        urllib.request.urlopen(req)
        print(f"✅ Created list view for {obj_name}")
    except urllib.error.HTTPError as e:
        print(f"❌ Failed to create list view for {obj_name}: {e}")
        # print(e.read()) # duplicate list view might fail?

def main():
    token = login()
    
    # Clean up bad object if exists
    delete_object(token, "TicketTicket")

    # Create Ticket Object
    create_object(token, "Ticket", "Tickets", "ticket")

    # Add Fields
    fields = [
        {"api_name": "priority", "label": "Priority", "type": "Picklist", "options": ["High", "Medium", "Low"], "required": True},
        {"api_name": "status", "label": "Status", "type": "Picklist", "options": ["New", "In Progress", "Closed"], "required": True},
        {"api_name": "description", "label": "Description", "type": "LongText"},
        {"api_name": "customer_id", "label": "Customer", "type": "Lookup", "reference_to": "account", "required": True},
        {"api_name": "contact_id", "label": "Contact", "type": "Lookup", "reference_to": "contact"}
    ]

    for f in fields:
        create_field(token, "ticket", f)

    create_list_view(token, "ticket")

    print("Phase 2 Setup Complete.")

if __name__ == "__main__":
    main()
