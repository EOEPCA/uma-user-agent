#!/usr/bin/env bash

ORIG_DIR="$(pwd)"
cd "$(dirname "$0")"
BIN_DIR="$(pwd)"

onExit() {
  docker-compose down
  cd "${ORIG_DIR}"
}

trap onExit EXIT

docker-compose up -d --build && docker-compose logs -f
