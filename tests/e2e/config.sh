#!/bin/bash
# tests/e2e/config.sh
# Centralized configuration for E2E tests

# API Configuration
export BASE_URL="${BASE_URL:-http://localhost:3001}"

# Test Credentials
export TEST_EMAIL="${TEST_EMAIL:-admin@test.com}"
export TEST_PASSWORD="${TEST_PASSWORD:-Admin123!}"

# Test State (preserved if already set by previous suite)
export TOKEN="${TOKEN:-}"
export USER_ID="${USER_ID:-}"
export CREATED_RECORD_ID="${CREATED_RECORD_ID:-}"

# Test Counters
export TOTAL_PASSED="${TOTAL_PASSED:-0}"
export TOTAL_FAILED="${TOTAL_FAILED:-0}"

# Colors for output
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export BLUE='\033[0;34m'
export CYAN='\033[0;36m'
export NC='\033[0m' # No Color

# Test Configuration
export TIMEOUT=30
export RETRY_COUNT=3
