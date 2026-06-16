#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

UPSTREAM="${UPSTREAM:-1.1.1.1,9.9.9.9,1.1.1.2}"
DNS_PORT="${DNS_PORT:-:5300}"
HTTP_PORT="${HTTP_PORT:-:8080}"
TIMEOUT="${TIMEOUT:-30s}"
CLEANUP="${CLEANUP:-45s}"
LOAD_DURATION="${DURATION:-10}"
LOAD_CLIENTS="${CLIENTS:-2}"
LOAD_QUERYFILE="${QUERYFILE:-load/queries.txt}"
LOAD_DNS_PORT="${DNS_PORT#:}"

command -v dnsperf > /dev/null || { echo ">> error: dnsperf is not installed"; exit 1; }

if [ ! -x ./fwdns ]; then
    echo ">> binary not found. compiling..."
    go build -o fwdns .
fi

echo ">> starting fwdns..."
./fwdns --dns "$DNS_PORT" --http "$HTTP_PORT" --upstream "$UPSTREAM" --timeout "$TIMEOUT" --cleanup "$CLEANUP" 2>&1 &
SERVER_PID=$!

cleanup() {
    echo ">> exiting..."
    kill "$SERVER_PID" 2> /dev/null || true
}
trap cleanup EXIT

sleep 1

echo ">> running load..."
dnsperf -s 127.0.0.1 -p "$LOAD_DNS_PORT" -d "$LOAD_QUERYFILE" -l "$LOAD_DURATION" -c "$LOAD_CLIENTS"
