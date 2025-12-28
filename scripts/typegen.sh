#!/bin/bash
# Generate constants for both backend and frontend from shared JSON
# Run from project root: ./scripts/typegen.sh

set -e

echo "🔄 Generating TypeScript constants..."
node scripts/generate-ts-constants.js

echo ""
echo "✅ All constants generated successfully!"
echo ""
echo "Generated files:"
echo "  - shared/generated/constants.ts"
echo "  - shared/generated/index.ts"
