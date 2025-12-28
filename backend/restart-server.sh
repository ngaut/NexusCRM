#!/bin/bash

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔄 NexusCRM Server Restart Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Step 1: Kill ALL processes on port 3001
echo "1. Killing all processes on port 3001..."
PIDS_ON_PORT=$(lsof -ti:3001 2>/dev/null || true)
if [ -n "$PIDS_ON_PORT" ]; then
    echo "   Found PIDs: $PIDS_ON_PORT"
    kill -9 $PIDS_ON_PORT 2>/dev/null || true
    echo "   ✅ Killed processes on port 3001"
else
    echo "   ℹ️  No processes found on port 3001"
fi

# Step 2: Wait for port to be freed
echo "2. Waiting for port 3001 to be freed..."
sleep 2
if lsof -ti:3001 >/dev/null 2>&1; then
    echo "   ❌ ERROR: Port 3001 still in use!"
    lsof -i:3001
    exit 1
fi
echo "   ✅ Port 3001 is free"

# Step 3: Rebuild server (optional - skip if binary is fresh)
if [ "${SKIP_BUILD}" != "1" ]; then
    echo "3. Building server..."
    cd /Users/qiliu/projects/nexuscrm/backend
    go build -o bin/server cmd/server/main.go
    if [ $? -ne 0 ]; then
        echo "   ❌ Build failed!"
        exit 1
    fi
    BUILD_TIME=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" bin/server)
    echo "   ✅ Build successful (${BUILD_TIME})"
else
    echo "3. Skipping build (SKIP_BUILD=1)"
fi

# Step 4: Load environment and start server
echo "4. Starting server..."
cd /Users/qiliu/projects/nexuscrm/backend

# BEST PRACTICE: Use set -a to auto-export all variables
# set -a enables automatic export of all variables
# source loads .env file
# Keep set -a active until AFTER server is started (this was the original bug)
set -a
source ../.env 2>/dev/null || {
    echo "   ❌ ERROR: ../.env not found"
    exit 1
}

# Start server in background BEFORE disabling auto-export
PORT=3001 ./bin/server > /tmp/server.log 2>&1 &
SERVER_PID=$!

# Now safe to disable auto-export
set +a

echo "   Server PID: $SERVER_PID"
echo "   Log file: /tmp/server.log"

# Step 5: Verify server started
echo "5. Verifying server startup..."

# Poll for health check (max 180 attempts * 2s = 360s = 6 minutes)
# Increased timeout for TiDB Cloud bootstrap which can be slow
MAX_ATTEMPTS=180
ATTEMPT=1
SERVER_READY=false

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    # Check if process is still running
    if ! ps -p $SERVER_PID > /dev/null; then
        echo "   ❌ Server process died!"
        echo "   Last 20 lines of log:"
        tail -20 /tmp/server.log
        exit 1
    fi

    # Check health endpoint
    HEALTH_CHECK=$(curl -s http://localhost:3001/health 2>&1 || echo "failed")
    if [[ "$HEALTH_CHECK" == *"golang"* ]]; then
        echo "   ✅ Server is healthy!"
        echo "   Response: $HEALTH_CHECK"
        SERVER_READY=true
        break
    fi

    echo "   ⏳ Waiting for server... ($ATTEMPT/$MAX_ATTEMPTS)"
    sleep 2
    ATTEMPT=$((ATTEMPT + 1))
done

if [ "$SERVER_READY" = false ]; then
    echo "   ❌ Health check failed after 6 minutes!"
    echo "   Last 20 lines of log:"
    tail -20 /tmp/server.log
    kill -9 $SERVER_PID
    exit 1
fi

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Server started successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "PID: $SERVER_PID"
echo "URL: http://localhost:3001"
echo "Log: tail -f /tmp/server.log"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo
