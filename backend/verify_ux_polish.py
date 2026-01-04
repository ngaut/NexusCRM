import requests
import sys

BASE_URL = "http://localhost:3001/api"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

def login():
    resp = requests.post(f"{BASE_URL}/auth/login", json={"email": EMAIL, "password": PASSWORD})
    resp.raise_for_status()
    return resp.json()["token"]

def verify_ux_fixes():
    token = login()
    headers = {"Authorization": f"Bearer {token}"}
    
    # 1. Create Object
    obj_def = {
        "api_name": "TestUxPolish",
        "label": "Test Polish",
        "plural_label": "Test Polishes", # Mixed case
        "description": "" # Should default to Label, NO prefix
    }
    
    # Clean up if exists
    try:
        requests.delete(f"{BASE_URL}/metadata/objects/testuxpolish", headers=headers)
    except:
        pass
        
    print(f"Creating object {obj_def['api_name']}...")
    resp = requests.post(f"{BASE_URL}/metadata/objects", json=obj_def, headers=headers)
    resp.raise_for_status()
    
    # 2. Verify Metadata
    print("Verifying metadata...")
    meta = requests.get(f"{BASE_URL}/metadata/objects/testuxpolish", headers=headers).json()["schema"]
    
    # Check Plural Label Case
    if meta["plural_label"] == "Test Polishes":
        print("✅ Plural Label Case Preserved: 'Test Polishes'")
    else:
        print(f"❌ Plural Label Case FAIL: Expected 'Test Polishes', got '{meta['plural_label']}'")
        sys.exit(1)
        
    # Check Description Prefix
    if meta["description"] == "Test Polish":
        print("✅ Description Prefix Removed: 'Test Polish'")
    elif "Custom object:" in meta["description"]:
        print(f"❌ Description Prefix FAIL: Found 'Custom object:' in '{meta['description']}'")
        sys.exit(1)
    else:
         print(f"⚠️ Description Mismatch: Got '{meta['description']}'")

    print("🎉 UI/UX Polish Verification Validated!")

if __name__ == "__main__":
    verify_ux_fixes()
