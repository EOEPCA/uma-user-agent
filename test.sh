#!/usr/bin/env bash

ORIG_DIR="$(pwd)"
cd "$(dirname "$0")"
BIN_DIR="$(pwd)"

onExit() {
  cd "${ORIG_DIR}"
}
trap onExit EXIT

GOFLAGS="-count=1" go test "$@" ./...
let status=$?
if test $status -eq 0; then
  echo SUCCESS
else
  echo FAILED
fi

exit $status
