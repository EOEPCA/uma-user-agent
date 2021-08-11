#!/usr/bin/env bash

ORIG_DIR="$(pwd)"
cd "$(dirname "$0")"
BIN_DIR="$(pwd)"

onExit() {
  cd "${ORIG_DIR}"
}
trap onExit EXIT

SUPPLIED_TAG="$1"
source ./container-info

docker-compose build
docker tag ${REPOSITORY}:latest ${REPOSITORY}:${TAG}

echo -e "\nCreated docker image: ${REPOSITORY}:latest => ${REPOSITORY}:${TAG}\n"

if test -n "${SUPPLIED_TAG}"; then
  echo "Pushing images..."
  docker push ${REPOSITORY}:latest
  docker push ${REPOSITORY}:${TAG}
fi
