
import http.client
import json

HOST = "localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

def login():
    conn = http.client.HTTPConnection(HOST)
    conn.request("POST", "/api/auth/login", json.dumps({"email": EMAIL, "password": PASSWORD}), {"Content-Type": "application/json"})
    return json.loads(conn.getresponse().read())["token"]

def verify():
    token = login()
    conn = http.client.HTTPConnection(HOST)
    headers = {"Authorization": f"Bearer {token}"}
    print("Checking test_large_v7 (Linear Fallback + Concurrency)...")
    conn.request("GET", "/api/metadata/objects/test_large_v7", headers=headers)
    resp = conn.getresponse()
    if resp.status == 200:
        data = json.load(resp)
        if "data" in data: data = data["data"]
        fields = {f['api_name']: f['type'] for f in data.get('fields', [])}
        val_type = fields.get('value', 'MISSING')
        if val_type == 'LongTextArea': print("PASS: test_large_v7 value is LongTextArea")
        else: print(f"FAIL: test_large_v7 value is {val_type}")
    else: print(f"FAIL: test_large_v7 error {resp.status}")

    conn.close()
    conn = http.client.HTTPConnection(HOST) # Reset
    
    print("Checking test_huge_v3 (Distributed Sampling + Concurrency)...")
    conn.request("GET", "/api/metadata/objects/test_huge_v3", headers=headers)
    resp = conn.getresponse()
    if resp.status == 200:
        data = json.load(resp)
        if "data" in data: data = data["data"]
        fields = {f['api_name']: f['type'] for f in data.get('fields', [])}
        val_type = fields.get('value', 'MISSING')
        if val_type == 'LongTextArea': print("PASS: test_huge_v3 value is LongTextArea")
        else: print(f"FAIL: test_huge_v3 value is {val_type}")
    else: print(f"FAIL: test_huge_v3 error {resp.status}")

    return # End here for now

    data = json.load(resp)
    if "data" in data: data = data["data"]
    fields = {f['api_name']: f for f in data.get('fields', [])}
    expectations = {
        "value": "LongTextArea"
    }
    
    for fname, expected in expectations.items():
        if fname not in fields:
            print(f"FAIL: {fname} missing")
            continue
            
        f = fields[fname]
        actual = f["type"]
        logical = f.get("logicalType")
        
        match = False
        if expected == "Lookup":
            if logical == "Lookup": match = True
            if actual == "VARCHAR(255)" and not logical: match = True # fallback
        elif expected == "Boolean":
             if actual in ("Boolean", "TINYINT(1)"): match = True
        elif expected == "Number":
             if actual in ("Number", "DECIMAL", "INT"): match = True
        elif expected == "DateTime":
             if actual in ("DateTime", "TIMESTAMP"): match = True
        elif expected == "Email":
             if actual == "Email": match = True
        elif expected == "Url":
             if actual == "Url": match = True
             
        if match:
            print(f"PASS: {fname} is {expected} ({actual})")
        else:
            print(f"FAIL: {fname} expected {expected}, got {actual}")

if __name__ == "__main__":
    verify()
