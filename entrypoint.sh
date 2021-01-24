#!/bin/bash

# This script is only intended for environment setup.

AIRTABLE_KEY_TEMP=$(gcloud secrets versions access 1 --secret="airtable-key")
if [ ! -z "$AIRTABLE_KEY_TEMP" ]; then
  export AIRTABLE_KEY=${AIRTABLE_KEY_TEMP}
fi

echo "Starting exporter service..."
./airtable-export $BUCKET_PATH
