#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."
. ./scripts/setup.sh

echo "Building image..."
echo
docker build -t airtable-export .

echo
echo "Running image..."
exec docker run \
	"${DOCKER_RUN_ARGS[@]}" \
	-p 8080:8080 \
	airtable-export \
	sh /entrypoint.sh server "$@"
