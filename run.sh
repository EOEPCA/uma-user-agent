#!/usr/bin/env bash

ORIG_DIR="$(pwd)"
cd "$(dirname "$0")"
BIN_DIR="$(pwd)"

onExit() {
  cd "${ORIG_DIR}"
}

trap onExit EXIT

go install github.com/cosmtrek/air@latest

export PORT=8080
export LOG_LEVEL=trace
export CLIENT_ID="22ba0c56-9780-4b0b-ad71-d745c166ca3b"
export CLIENT_SECRET="0e3e1d0d-9002-4d44-bbff-a170efa18512"
export CONFIG_DIR="./"

air
