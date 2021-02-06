#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."
. ./scripts/setup.sh

echo "Building image..."
echo
docker build -t freshcf . -f monitoring/freshcf/Dockerfile

echo
echo "Running image..."
exec docker run \
	-e "AIRTABLE_KEY=$AIRTABLE_KEY" \
	-e "HONEYCOMB_KEY=$HONEYCOMB_KEY" \
	-e "TESTING_BUCKET=$TESTING_BUCKET" \
	${GOOGLE_AUTH_BIND:+'-v' "$GOOGLE_AUTH_BIND"} \
	--rm \
	-p 8080:8080 \
	freshcf
