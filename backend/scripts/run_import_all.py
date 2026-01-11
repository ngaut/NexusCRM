
import os
import subprocess
import json
import time
import http.client
import sys

HOST = "localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"
DATA_DIR = "../../salesforce_data"


# Order matters for foreign keys!
PRIORITY_ORDER = [
    "user",
    "user_role",
    "profile",
    "record_type",
    "account",
    "contact", 
    "lead",
    "campaign",
    "opportunity",
    "case",
    "task", 
    "event",
    "note"
]

def request(method, path, body=None):
    conn = http.client.HTTPConnection(HOST)
    headers = {"Content-Type": "application/json"}
    
    json_body = json.dumps(body) if body else None
    conn.request(method, path, body=json_body, headers=headers)
    resp = conn.getresponse()
    data = resp.read()
    conn.close()
    
    if resp.status >= 400:
        print(f"Error {resp.status}: {data.decode('utf-8')}")
        return None
        
    return json.loads(data)

def login():
    data = request("POST", "/api/auth/login", {"email": EMAIL, "password": PASSWORD})
    if data:
        return data["token"]
    return None

def run_import():
    token = login()
    if not token:
        print("Login failed")
        return

    data_abs_path = os.path.abspath(os.path.join(os.path.dirname(__file__), DATA_DIR))

    # Walk directory to find items (CSV files or Parquet Directories)
    data_items = []

    # get immediate children
    for name in os.listdir(data_abs_path):
        full_path = os.path.join(data_abs_path, name)
        
        # CSV File
        if name.endswith(".csv"):
             data_items.append(name)
        # Directory (Parquet?)
        elif os.path.isdir(full_path):
            # Check if likely parquet
            # Just assume any directory in 'data' is a partitioned table
            if not name.startswith('.'):
                data_items.append(name)
    
    # Sort
    sorted_items = []
    
    # 1. Priority
    for p in PRIORITY_ORDER:
        for f in data_items:
            # f might be "account" or "account.csv"
            clean_name = f.replace(".csv", "").lower()
            if clean_name == p or clean_name == f"salesforce__{p}" or clean_name == f"stg_salesforce__{p}":
                if f not in sorted_items:
                    sorted_items.append(f)
                    
    # 2. Others
    for f in data_items:
        if f not in sorted_items:
            sorted_items.append(f)

    print(f"Found {len(sorted_items)} items. Sorted order: {sorted_items[:5]}...")

    for item_name in sorted_items:
        # Normalize Object Name
        # stg_salesforce__user -> user
        # salesforce__user -> user
        # user -> user
        
        lower = item_name.lower().replace(".csv", "")
        if lower.startswith("stg_salesforce__"):
            obj_name = "stg_" + lower[16:]
        elif lower.startswith("salesforce__"):
            obj_name = lower[12:]
        else:
            obj_name = lower

        file_path = os.path.join(data_abs_path, item_name)
        
        print(f"\n>>> Importing {obj_name} from {item_name}...")
        
        # Use venv python if available
        python_bin = "python3"
        venv_python = os.path.join(os.path.dirname(__file__), "../venv/bin/python")
        if os.path.exists(venv_python):
            python_bin = venv_python
        
        cmd = [
            python_bin, "-m", "scripts.migration_tool.main",
            "--file", file_path,
            "--obj", obj_name,
            "--concurrency", "30",
            "--url", f"http://{HOST}",
            "--token", token
        ]
        
        try:
            subprocess.run(cmd, check=True)
        except subprocess.CalledProcessError as e:
            print(f"!!! Error importing {item_name}: {e}")
            time.sleep(1)

def clean_checkpoints():
    print("🧹 Cleaning old checkpoints and logs...")
    # Remove import_stats.csv if exists
    if os.path.exists("import_stats.csv"):
        os.remove("import_stats.csv")
    
    # Remove any local .checkpoint files
    for f in os.listdir("."):
        if f.endswith(".checkpoint") or f.endswith(".log"):
            try:
                os.remove(f)
            except Exception:
                pass

if __name__ == "__main__":
    clean_checkpoints()
    run_import()
