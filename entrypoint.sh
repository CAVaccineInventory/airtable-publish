#!/bin/sh

AIRTABLE_KEY_TEMP=$(gcloud secrets versions access 1 --secret="airtable-key" 2>/dev/null)
if [ -n "$AIRTABLE_KEY_TEMP" ]; then
	export AIRTABLE_KEY=${AIRTABLE_KEY_TEMP}
fi

HONEYCOMB_KEY_TEMP=$(gcloud secrets versions access 1 --secret="honeycomb-key" 2>/dev/null)
if [ -n "$HONEYCOMB_KEY_TEMP" ]; then
	export HONEYCOMB_KEY=${HONEYCOMB_KEY_TEMP}
fi

if [ -f /testing-key.json ]; then
	export GOOGLE_APPLICATION_CREDENTIALS=/testing-key.json
elif [ -d /testing-key.json ]; then
	echo
	echo "Testing file specified, but not found on your host."
	echo
	echo "Check the path to your testing key that you passed as"
	echo "the first part to -v; it has been created as a directory."
	exit 1
fi

COMMAND="${1:-server}"
shift
if [ ! -x "/$COMMAND" ]; then
	echo "Unknown command: $COMMAND"
	exit 1
fi
exec "/$COMMAND" "$@"
