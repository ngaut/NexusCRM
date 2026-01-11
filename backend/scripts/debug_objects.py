
import http.client
import json

HOST = "localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

def request(method, path, body=None, token=None):
    conn = http.client.HTTPConnection(HOST)
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    
    conn.request(method, path, body=json.dumps(body) if body else None, headers=headers)
    resp = conn.getresponse()
    data = resp.read()
    conn.close()
    return resp.status, json.loads(data)

def login():
    status, data = request("POST", "/api/auth/login", {"email": EMAIL, "password": PASSWORD})
    return data["token"]

def list_objects():
    token = login()
    print("Listing objects...")
    status, data = request("GET", "/api/metadata/objects", token=token)
    
    if status == 200:
        if "data" in data:
            objs = data["data"]
        else:
            objs = data
            
        names = [o["api_name"] for o in objs]
        print(f"Found {len(names)} objects.")
        print(f"Names: {names}")
        
        if "user" in names:
            print("user exists.")
        if "account" in names:
            print("account exists.")
        if "stg_user" in names:
            print("stg_user exists.")
    else:
        print(f"Error {status}: {data}")

if __name__ == "__main__":
    list_objects()
