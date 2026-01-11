
import sys
import os
import json

# Add parent directory to path to import migration_tool modules
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from backend.scripts.migration_tool.client import NexusClient

# Hardcoded credentials (same as run_import_all.py)
HOST = "localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

def login():
    client = NexusClient(f"http://{HOST}", "")
    data, err = client.request("POST", "/api/auth/login", {"email": EMAIL, "password": PASSWORD})
    if err:
        print(f"Login failed: {err}")
        return None
    return data["token"]

def list_apps():
    token = login()
    if not token: return
    
    client = NexusClient(f"http://{HOST}", token)
    data, err = client.request("GET", "/api/metadata/apps")
    
    if err:
        print(f"Error fetching apps: {err}")
        return

    print(json.dumps(data, indent=2))
    
    # Check for "data platform" specifically
    found = False
    if "data" in data:
        for app in data["data"]:
            if app.get("name").lower() == "data platform" or app.get("label").lower() == "data platform":
                print(f"\nFOUND APP: {app['id']} ({app['label']})")
                found = True
    
    if not found:
        print("\n'data platform' app NOT found.")

if __name__ == "__main__":
    list_apps()
