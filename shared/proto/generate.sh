#!/bin/bash
# Generate constants from protobuf for Go and TypeScript
# Prerequisites: protoc, protoc-gen-go, protoc-gen-ts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "📦 Generating constants from protobuf..."

# Check for protoc
if ! command -v protoc &> /dev/null; then
    echo "❌ protoc not found. Install with: brew install protobuf"
    exit 1
fi

# Generate Go code
echo "   🔧 Generating Go code..."
mkdir -p "$PROJECT_ROOT/backend/pkg/constants/pb"

protoc \
    --proto_path="$SCRIPT_DIR" \
    --go_out="$PROJECT_ROOT/backend/pkg/constants/pb" \
    --go_opt=paths=source_relative \
    "$SCRIPT_DIR/constants.proto"

if [ $? -eq 0 ]; then
    echo "   ✅ Go code generated: backend/pkg/constants/pb/constants.pb.go"
else
    echo "   ❌ Go generation failed!"
    exit 1
fi

# Generate TypeScript code (optional - requires ts-proto or similar)
if command -v protoc-gen-ts_proto &> /dev/null; then
    echo "   🔧 Generating TypeScript code..."
    mkdir -p "$PROJECT_ROOT/frontend/src/generated"
    
    protoc \
        --proto_path="$SCRIPT_DIR" \
        --ts_proto_out="$PROJECT_ROOT/frontend/src/generated" \
        --ts_proto_opt=esModuleInterop=true \
        --ts_proto_opt=outputEncodeMethods=false \
        --ts_proto_opt=outputClientImpl=false \
        "$SCRIPT_DIR/constants.proto"
    
    if [ $? -eq 0 ]; then
        echo "   ✅ TypeScript code generated: frontend/src/generated/constants.ts"
    else
        echo "   ⚠️  TypeScript generation failed (optional)"
    fi
else
    echo "   ⚠️  protoc-gen-ts_proto not found - skipping TypeScript generation"
    echo "      Install with: npm install -g ts-proto"
fi

echo ""
echo "✅ Generation complete!"
echo ""
echo "To use in Go:"
echo "   import pb \"github.com/nexuscrm/backend/pkg/constants/pb\""
echo "   trigger := pb.TriggerType_TRIGGER_TYPE_BEFORE_CREATE"
echo ""
