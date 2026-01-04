#!/bin/bash
# reseed.sh - Tear down and re-seed metadata

# Source API helpers
source tests/e2e/lib/api.sh

echo "🔄 Logging in..."
if ! api_login "admin@test.com" "Admin123!"; then
    echo "❌ Login failed"
    exit 1
fi

echo "🔄 Tearing down existing metadata..."
# Temporarily disable exit on error for teardown
# Temporarily disable exit on error for teardown
set +e
source tests/e2e/lib/teardown.sh
set -e

echo "🔄 Re-seeding metadata with fixes..."
source tests/e2e/lib/setup.sh
ensure_test_objects

echo "✅ Reseed complete!"
