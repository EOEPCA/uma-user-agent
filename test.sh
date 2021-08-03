#!/usr/bin/env bash

ORIG_DIR="$(pwd)"
cd "$(dirname "$0")"
BIN_DIR="$(pwd)"

onExit() {
  cd "${ORIG_DIR}"
}

trap onExit EXIT

export LOG_LEVEL=trace
export CLIENT_ID="691892fd-1e58-4e44-8355-1f3b7634af8f"
export CLIENT_SECRET="eb28831f-3e0f-4580-a103-8fd1e0adbb3c"

go test -v ./test/... 
# go test -v ./test/uma/uma-client_test.go
