#!/bin/bash

# This script is only intended for environment setup.

AIRTABLE_KEY_TEMP=$(gcloud secrets versions access 1 --secret="airtable-key" 2>/dev/null)
if [ -n "$AIRTABLE_KEY_TEMP" ]; then
	export AIRTABLE_KEY=${AIRTABLE_KEY_TEMP}
fi

if [ -f /testing-key.json ]; then
	export GOOGLE_APPLICATION_CREDENTIALS=/testing-key.json
	gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS
elif [ -d /testing-key.json ]; then
	echo
	echo "Testing file specified, but not found on your host."
	echo
	echo "Check the path to your testing key that you passed as"
	echo "the first part to -v; it has been created as a directory."
	exit 1
fi

echo "Starting exporter service..."
./pipeline
