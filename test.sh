#!/usr/bin/env bash

ORIG_DIR="$(pwd)"
cd "$(dirname "$0")"
BIN_DIR="$(pwd)"

onExit() {
  cd "${ORIG_DIR}"
}

trap onExit EXIT

export LOG_LEVEL=trace
export CLIENT_ID="528e4b49-1f6f-4ff7-bc61-36d08b00c69a"
export CLIENT_SECRET="8afaadc0-d09b-4b28-bedc-765dcf5c27f4"

GOFLAGS="-count=1" go test "$@" ./...
let status=$?
if test $status -eq 0; then
  echo SUCCESS
else
  echo FAILED
fi

exit $status
