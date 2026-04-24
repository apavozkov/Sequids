#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v mosquitto_pub >/dev/null 2>&1; then
  echo "ERROR: mosquitto_pub not found in PATH."
  echo "Install it on Ubuntu: sudo apt update && sudo apt install -y mosquitto-clients"
  exit 1
fi

./scripts/stop_demo.sh || true

go run ./cmd/central serve -grpc-addr :50051 -metrics-addr :8080 -db ./sequids.db -formulas ./configs/formulas/formulas.yaml -anomalies ./configs/anomalies/anomalies.yaml &
CENTRAL_PID=$!
sleep 1

go run ./cmd/worker -grpc-addr :50052 -metrics-addr :8090 -central-grpc 127.0.0.1:50051 -mqtt-host localhost -mqtt-port 1883 -influx-url http://localhost:8086 -influx-token sequids-token -influx-org sequids -influx-bucket metrics &
WORKER_PID=$!

cleanup() {
  kill "$CENTRAL_PID" "$WORKER_PID" 2>/dev/null || true
}
trap cleanup EXIT

sleep 2
go run ./cmd/sequidsctl start -grpc 127.0.0.1:50051 -scenario-file ./examples/greenhouse.dsl -scenario-name greenhouse -seed 42

wait
