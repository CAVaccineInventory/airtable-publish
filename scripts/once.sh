#!/usr/bin/env bash

set -eux

cd "$(dirname "$0")/.."
. ./scripts/setup.sh

echo "Building image..."
echo
docker build --build-arg COMMIT_SHA="$COMMIT_SHA" -t airtable-export .

echo
echo "Running image..."
exec docker run \
	"${DOCKER_RUN_ARGS[@]}" \
	airtable-export \
	sh /entrypoint.sh once "$@"
