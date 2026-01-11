
import os
import subprocess
import json
import time
import http.client
import sys

HOST = "localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"
DATA_DIR = "../../salesforce_data/data"
IMPORT_SCRIPT = "migration/import_salesforce.py"

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

def get_csv_files():
    data_path = os.path.join(os.path.dirname(__file__), DATA_DIR)
    if not os.path.exists(data_path):
        print(f"Data dir not found: {data_path}")
        return []
    
    files = []
    for f in os.listdir(data_path):
        if f.endswith(".csv"):
            files.append(f)
    return files

def run_import():
    token = login()
    if not token:
        print("Login failed")
        return

    all_files = get_csv_files()
    if not all_files:
        print("No CSV files found.")
        return

    # Sort files
    sorted_files = []
    
    # 1. Priority files
    for p in PRIORITY_ORDER:
        # Check for exact Match or variations
        for f in all_files:
            name = f.lower().replace(".csv", "")
            # Check exact match or standard prefixes like "salesforce__user"
            if name == p or name == f"salesforce__{p}" or name == f"stg_salesforce__{p}":
                if f not in sorted_files:
                    sorted_files.append(f)
    
    # 2. Rest of files
    for f in all_files:
        if f not in sorted_files:
            sorted_files.append(f)

    print(f"Found {len(all_files)} files. Sorted order: {sorted_files[:5]}...")

    script_path = os.path.join(os.path.dirname(__file__), IMPORT_SCRIPT)
    data_abs_path = os.path.abspath(os.path.join(os.path.dirname(__file__), DATA_DIR))

    for filename in sorted_files:
        # Normalize: 
        # stg_salesforce__user.csv -> stg_user
        # salesforce__user.csv -> user
        # user.csv -> user
        
        name = filename.lower()
        if name.endswith(".csv"):
            name = name[:-4]
            
        if name.startswith("stg_salesforce__"):
            obj_name = "stg_" + name[16:]
        elif name.startswith("salesforce__"):
            obj_name = name[12:]
        else:
            obj_name = name

        file_path = os.path.join(data_abs_path, filename)
        # Run import script
        # script_path = os.path.join(os.path.dirname(__file__), "migration/import_salesforce.py")
        # Run migration tool module
        # script_path = os.path.join(os.path.dirname(__file__), "import_generic.py")
        
        print(f"\n>>> Importing {obj_name} from {filename}...")
        
        cmd = [
            "python3", "-m", "scripts.migration_tool.main",
            "--file", file_path,
            "--obj", obj_name,
            "--concurrency", "30",
            "--url", f"http://{HOST}",
            "--token", token
        ]
        
        try:
            # We want to see output live, so we don't capture it (allow stdout)
            subprocess.run(cmd, check=True)
        except subprocess.CalledProcessError as e:
            print(f"!!! Error importing {filename}: {e}")
            # Continue to next file? Yes, usually.
            time.sleep(1)

def clean_checkpoints():
    print("🧹 Cleaning old checkpoints and logs...")
    # Remove import_stats.csv if exists
    if os.path.exists("import_stats.csv"):
        os.remove("import_stats.csv")
    # Clean checkpoints dir if we had one (currently using file tracking logic in main?)
    # The current importer doesn't strictly use a checkpoints dir, but we should ensure fresh start.
    pass

if __name__ == "__main__":
    clean_checkpoints()
    run_import()
