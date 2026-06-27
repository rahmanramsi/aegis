#!/usr/bin/env bash
# Smoke test: full user flow from register to agent creation.
# Requires: bin/aegisd and bin/aegis-agent binaries in PATH or current dir.
set -euo pipefail

DIR="$(cd "$(dirname "$0")/.." && pwd)"
GATEWAY="${DIR}/bin/aegisd"
AGENT_BIN="${DIR}/bin/aegis-agent"
DB_PATH="$(mktemp -d)/smoke.db"
PORT=19876

cleanup() {
    kill %1 2>/dev/null || true
    kill %2 2>/dev/null || true
    rm -rf "$(dirname "$DB_PATH")"
}
trap cleanup EXIT

echo "=== Building ==="
cd "$DIR"
go build -o "$GATEWAY" ./cmd/aegisd
go build -o "$AGENT_BIN" ./cmd/aegis-agent

echo "=== Starting gateway ==="
AEGIS_DATABASE_URL="$DB_PATH" AEGIS_ADDR=":$PORT" "$GATEWAY" &
sleep 2

FAIL=0
check() {
    local label="$1" expected="$2" actual="$3"
    if [ "$actual" != "$expected" ]; then
        echo "  FAIL $label: expected '$expected', got '$actual'"
        FAIL=1
    else
        echo "  PASS $label"
    fi
}

wait_daemon_status() {
    local want="$1"
    local status=""
    for _ in {1..20}; do
        status=$(curl -sf -H "Authorization: Bearer $NEWKEY" "$BASE/api/v1/daemons" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['status'])") || true
        if [ "$status" = "$want" ]; then
            echo "$status"
            return 0
        fi
        sleep 0.5
    done
    echo "$status"
}

BASE="http://localhost:$PORT"

# 1. Health
echo "--- Health ---"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/health")
check "health" "200" "$CODE"

# 2. Register
echo "--- Register ---"
REG=$(curl -sf -X POST "$BASE/api/v1/auth/register" \
    -H 'Content-Type: application/json' \
    -d '{"email":"smoke@test.dev","password":"smoke123"}')
KEY=$(echo "$REG" | python3 -c "import sys,json; print(json.load(sys.stdin)['api_key'])")
check "register" "0" "$(if [ -n "$KEY" ]; then echo 0; else echo 1; fi)"

# 3. Login
echo "--- Login ---"
LOGIN=$(curl -sf -X POST "$BASE/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d '{"email":"smoke@test.dev","password":"smoke123"}')
NEWKEY=$(echo "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['api_key'])")
check "login" "0" "$(if [ -n "$NEWKEY" ]; then echo 0; else echo 1; fi)"

# 4. Unauthenticated request
echo "--- Auth required ---"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/workspaces")
check "no-auth" "401" "$CODE"

# 5. Create workspace
echo "--- Create workspace ---"
WS=$(curl -sf -X POST "$BASE/api/v1/workspaces" \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $NEWKEY" \
    -d '{"name":"Smoke","slug":"smoke"}')
WSID=$(echo "$WS" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
check "create-ws" "0" "$(if [ -n "$WSID" ]; then echo 0; else echo 1; fi)"

# 6. List workspaces (registration creates a default workspace; this test adds one more)
echo "--- List workspaces ---"
COUNT=$(curl -sf -H "Authorization: Bearer $NEWKEY" "$BASE/api/v1/workspaces" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))")
check "list-ws" "2" "$COUNT"

# 7. Create daemon
echo "--- Create daemon ---"
DM=$(curl -sf -X POST "$BASE/api/v1/daemons" \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $NEWKEY" \
    -d '{"name":"smoke-daemon"}')
DMID=$(echo "$DM" | python3 -c "import sys,json; print(json.load(sys.stdin)['daemon']['id'])")
DTOKEN=$(echo "$DM" | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
check "create-daemon" "0" "$(if [ -n "$DMID" ]; then echo 0; else echo 1; fi)"

# 8. Start daemon
echo "--- Start daemon ---"
PATH="/usr/bin:/bin" AEGIS_DAEMON_ID="$DMID" AEGIS_DAEMON_TOKEN="$DTOKEN" AEGIS_GATEWAY_URL="ws://localhost:$PORT/ws/daemon" "$AGENT_BIN" &

STATUS=$(wait_daemon_status "online")
check "daemon-online" "online" "$STATUS"

# 9. Create agent
echo "--- Create agent ---"
AGENT=$(curl -sf -X POST "$BASE/api/v1/workspaces/$WSID/agents" \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $NEWKEY" \
    -d '{"name":"smoke-agent","daemon_id":"'"$DMID"'","harness":"echo","personality":"Be helpful."}')
ANAME=$(echo "$AGENT" | python3 -c "import sys,json; print(json.load(sys.stdin)['agent']['name'])")
check "create-agent" "smoke-agent" "$ANAME"

echo ""
if [ "$FAIL" -eq 0 ]; then
    echo "═══════════════════════════"
    echo "  SMOKE TEST: ALL PASSED"
    echo "═══════════════════════════"
else
    echo "═══════════════════════════"
    echo "  SMOKE TEST: FAILURES"
    echo "═══════════════════════════"
    exit 1
fi
