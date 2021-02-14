#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."
if [ ! -f .env ]; then
	cat <<EOF

No .env found; I have created a bare-bones one which includes
instructions on finding the relevant keys.

EOF
	cat >.env <<EOF
# -*-sh-*-
# This file defines environment variables used for running things
# locally.

# Airtable key, to fetch data; can be found at https://airtable.com/account
# If you need access to the Airtable itself, see the link in the topic
# of the #phone-banking channel on Discord.  Required.

AIRTABLE_KEY=

# Honeycomb API key, to test uploading spans and metrics; can be found
# at https://ui.honeycomb.io/teams/vaccinateca/. Required if reporting
# metrics, which default to off in 'once', and can be disabled with
# '-metrics=false' for pipeline server.

HONEYCOMB_KEY=

# Optional Google Cloud authentication gets stored in
# 'testing-key.json', to test uploading files and metrics; see
# 'pipeline/README.md' for instructions.
EOF
fi

. .env

if [ "$AIRTABLE_KEY" = "" ]; then
	echo "Fetching cannot work without an AIRTABLE_KEY set; update your .env"
	exit 1
fi

GOOGLE_AUTH_BIND=
if [ -f "testing-key.json" ]; then
	GOOGLE_AUTH_BIND="$(pwd)/testing-key.json:/testing-key.json"
fi

# We clean out `local/` before every run.
rm -rf local/
LOCAL_BIND="$(pwd)/local/:/src/local/"

DOCKER_RUN_ARGS=(
	--rm
	-e "AIRTABLE_KEY=${AIRTABLE_KEY}"
	-e "HONEYCOMB_KEY=${HONEYCOMB_KEY}"
	${GOOGLE_AUTH_BIND:+ -v "$GOOGLE_AUTH_BIND"}
	${LOCAL_BIND:+ -v "$LOCAL_BIND"}
)
