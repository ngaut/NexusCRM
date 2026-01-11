
import sys
import os
import json

sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from backend.scripts.migration_tool.client import NexusClient

HOST = "localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

OBJECTS = [
    "elvance_opp_scorecard_c",          # CORRECTION: elvance_ NOT pqcrush_
    "pqcrush_relationship_map_c", 
    "pqcrush_relationship_map_member_c",
    "elvance_elvance_risk_entry_c",
    "elvance_activity_log_c",
    "opportunity",
    "global_arr_forecast_c"
]

def login():
    client = NexusClient(f"http://{HOST}", "")
    data, err = client.request("POST", "/api/auth/login", {"email": EMAIL, "password": PASSWORD})
    if err:
        print(f"Login failed: {err}")
        return None
    return data["token"]

def analyze_objects():
    token = login()
    if not token: return
    client = NexusClient(f"http://{HOST}", token)
    
    print("="*60)
    print("ANALYSIS: Data Platform & MEDDPIC Potential")
    print("="*60)

    for obj in OBJECTS:
        print(f"\n🔍 Analyzing Object: {obj}")
        
        query = {
            "object_api_name": obj,
            "limit": 5,
            "offset": 0
        }
        res, err = client.request("POST", "/api/data/query", query)
        
        if err:
            print(f"❌ Error querying {obj}: {err}")
            continue
            
        # Robust Response Handling
        data = []
        total = "Unknown"
        
        raw_data = res.get("data")
        if isinstance(raw_data, list):
            data = raw_data
            total = len(data)
        elif isinstance(raw_data, dict):
            data = raw_data.get("records", [])
            total = raw_data.get("total", 0)
        
        print(f"   Total Records: {total}")
        
        if len(data) > 0:
            print("   Sample Keys & Values:")
            first_rec = data[0]
            keys = sorted(first_rec.keys())
            
            # Print interesting fields (looking for MEDDPIC clues)
            for k in keys:
                 if k.startswith("__") or k in ["id", "created_date", "created_by_id", "last_modified_date", "last_modified_by_id", "owner_id", "is_deleted", "system_modstamp"]: continue
                 
                 val_str = str(first_rec[k])
                 if len(val_str) > 100: val_str = val_str[:100] + "..."
                 print(f"    - {k}: {val_str}")
        else:
            print("   (Empty Table or No Data Fetched)")

if __name__ == "__main__":
    analyze_objects()
