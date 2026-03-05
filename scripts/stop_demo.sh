#!/usr/bin/env bash
set -euo pipefail

# Stops stray Sequids central/worker processes started via go run or built binaries.
pkill -f "cmd/central serve" 2>/dev/null || true
pkill -f "cmd/worker" 2>/dev/null || true
pkill -f "/central serve" 2>/dev/null || true
pkill -f "/worker" 2>/dev/null || true

echo "Requested stop for running Sequids central/worker processes."
