#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

go run ./cmd/central serve -rpc-addr :50051 -metrics-addr :8080 -db ./sequids.db -formulas ./configs/formulas/formulas.yaml -anomalies ./configs/anomalies/anomalies.yaml &
CENTRAL_PID=$!
sleep 1

go run ./cmd/worker -rpc-addr :50052 -metrics-addr :8090 -central-rpc 127.0.0.1:50051 -mqtt-host localhost -mqtt-port 1883 -influx-url http://localhost:8086 -influx-token sequids-token -influx-org sequids -influx-bucket metrics &
WORKER_PID=$!

cleanup() {
  kill "$CENTRAL_PID" "$WORKER_PID" 2>/dev/null || true
}
trap cleanup EXIT

sleep 2
SCENARIO_RESPONSE=$(go run ./cmd/central push-scenario -rpc 127.0.0.1:50051 -file ./examples/greenhouse.dsl -name greenhouse)
SCENARIO_ID=$(echo "$SCENARIO_RESPONSE" | sed -E 's/.*"scenario_id":"([^"]+)".*/\1/')
go run ./cmd/central run -rpc 127.0.0.1:50051 -scenario "$SCENARIO_ID" -seed 42

wait
