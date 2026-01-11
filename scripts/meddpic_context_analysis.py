
import sys
import os
import json
from collections import defaultdict

sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from backend.scripts.migration_tool.client import NexusClient

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

def get_records_safe(res):
    """Helper to safely extract records from API response data which can be list or dict."""
    if not res: return [], 0
    raw_data = res.get("data")
    if isinstance(raw_data, list):
        return raw_data, len(raw_data)
    elif isinstance(raw_data, dict):
        return raw_data.get("records", []), raw_data.get("total", 0)
    return [], 0

def analyze_business_logic():
    token = login()
    if not token: return
    client = NexusClient(f"http://{HOST}", token)
    
    print("="*60)
    print("MEDDPIC BUSINESS ANALYSIS (DEBUG)")
    print("="*60)

    # 1. Distinct Roles
    print("\n1. 🔍 Analyzing Relationship Map Roles...")
    query = {
        "object_api_name": "pqcrush_relationship_map_member_c",
        "limit": 100, # Reduced for debug speed
        "offset": 0
    }
    res, err = client.request("POST", "/api/data/query", query)
    if not err:
        data, _ = get_records_safe(res)
        print(f"   Fetched {len(data)} members.")
        if len(data) > 0:
            print(f"   Sample Member 0: {json.dumps(data[0], default=str)[:200]}")
            
    # 2. Scorecard Linkage
    print("\n2. 🔍 Analyzing Scorecard Linkages...")
    
    q_sc_full = {
        "object_api_name": "elvance_opp_scorecard_c", 
        "limit": 10, # Analyze first 10 strictly
        "offset": 0
    }
    r_sc_full, err = client.request("POST", "/api/data/query", q_sc_full)
    
    if not err:
        scorecards, _ = get_records_safe(r_sc_full)
        print(f"   Fetched {len(scorecards)} scorecards.")
        
        for i, sc in enumerate(scorecards):
            opp_id = sc.get("opportunity_c")
            print(f"   [{i}] Scorecard ID: {sc.get('id')} -> Opp ID: {opp_id}")
            
            if not opp_id: 
                print("      ⚠️ No Opportunity ID linked.")
                continue
            
            # Get Opp
            ropp, err = client.request("GET", f"/api/data/opportunity/{opp_id}")
            if err:
                print(f"      ❌ Error fetching Opp: {err}")
                continue
            
            opp = ropp.get("data")
            if not opp:
                print("      ⚠️ Opp not found (404/Empty).")
                continue
                
            stage = str(opp.get("stage_name", ""))
            amount = str(opp.get("amount", "0"))
            print(f"      ✅ Linked to Opp: {opp.get('name')} | Stage: {stage} | Amt: {amount}")

if __name__ == "__main__":
    analyze_business_logic()
