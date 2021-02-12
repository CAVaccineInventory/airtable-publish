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
	-e "AIRTABLE_KEY=$AIRTABLE_KEY" \
	-e "HONEYCOMB_KEY=$HONEYCOMB_KEY" \
	${GOOGLE_AUTH_BIND:+'-v' "$GOOGLE_AUTH_BIND"} \
	${LOCAL_BIND:+'-v' "$LOCAL_BIND"} \
	--rm \
	-p 8080:8080 \
	airtable-export \
	sh /entrypoint.sh server "$@"
