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
# of the #phone-bankingi channel on Discord.
AIRTABLE_KEY=

# Honeycomb API key, to test uploading spans and metrics; can be found
# at https://ui.honeycomb.io/teams/vaccinateca
HONEYCOMB_KEY=

# Google Cloud authentication gets stored in 'testing-key.json', to test
# uploading files and metrics; see 'pipeline/README.md' for
# instructions.
TESTING_BUCKET=
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
