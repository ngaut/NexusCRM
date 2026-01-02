import json
import urllib.request
import urllib.error
import sys

BASE_URL = "http://localhost:3001"
EMAIL = "admin@test.com"
PASSWORD = "Admin123!"

def login():
    print(f"Logging in as {EMAIL}...")
    data = json.dumps({"email": EMAIL, "password": PASSWORD}).encode('utf-8')
    req = urllib.request.Request(
        f"{BASE_URL}/api/auth/login",
        data=data,
        headers={'Content-Type': 'application/json'}
    )
    
    try:
        with urllib.request.urlopen(req) as response:
            resp_body = response.read().decode('utf-8')
            resp_json = json.loads(resp_body)
            if resp_json.get("success"):
                return resp_json.get("token")
            else:
                print(f"Login failed: {resp_body}")
                sys.exit(1)
    except urllib.error.URLError as e:
        print(f"Connection error: {e}")
        sys.exit(1)

def test_thinking(token):
    print("\nSending chat request: 'add navigator items'...")
    data = json.dumps({
        "messages": [{"role": "user", "content": "add navigator items"}]
    }).encode('utf-8')
    
    req = urllib.request.Request(
        f"{BASE_URL}/api/agent/chat/stream",
        data=data,
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        }
    )

    thinking_received = False
    
    try:
        with urllib.request.urlopen(req) as response:
            print("\n--- Event Stream ---")
            for line in response:
                line = line.decode('utf-8').strip()
                if line.startswith("data:"):
                    data_str = line[5:].strip()
                    try:
                        event = json.loads(data_str)
                        ev_type = event.get("type")
                        
                        if ev_type == "thinking":
                            thinking_received = True
                            print(f"🟣 THINKING: {event.get('content')}")
                        elif ev_type == "tool_call":
                            print(f"🛠️  TOOL: {event.get('tool_name')} ({event.get('tool_args')})")
                        elif ev_type == "tool_result":
                            print(f"✅ RESULT: {event.get('tool_name')}")
                        elif ev_type == "content":
                            print(f"📝 CONTENT: {event.get('content')}")
                        elif ev_type == "error":
                            print(f"❌ ERROR: {event.get('content')}")
                        elif ev_type == "done":
                            print("🏁 DONE")
                    except json.JSONDecodeError:
                        pass
    except urllib.error.URLError as e:
        print(f"Stream error: {e}")

    print("\n--- Summary ---")
    if thinking_received:
        print("✅ SUCCESS: 'thinking' event received!")
    else:
        print("⚠️  WARNING: No 'thinking' event received.")
        print("Most likely the current LLM model does not support returning content with tool calls.")
        print("To fix, either:")
        print("1. Use a smarter model (GPT-4 / Claude)")
        print("2. Prompt engineer the model to 'think out loud first'")

if __name__ == "__main__":
    try:
        token = login()
        test_thinking(token)
    except KeyboardInterrupt:
        print("\nCancelled.")
