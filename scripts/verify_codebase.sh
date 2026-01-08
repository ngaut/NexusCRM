#!/bin/bash
# verify_codebase.sh
# Verifies the health of both Backend and Frontend.

set -e

echo "🔍 Verifying Backend..."
cd backend
go vet ./...
# Ensure codegen is clean (optional but good)
# go run ./cmd/codegen

echo "✅ Backend Verified."

echo "🔍 Verifying Frontend..."
cd ../frontend
# Run Type Check (tsc)
# npm run lint runs 'tsc --noEmit' per package.json
npm run lint

echo "✅ Frontend Verified."

echo "🎉 Codebase Verification Complete!"
