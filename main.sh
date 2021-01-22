#!/bin/bash

# Being sunset, we never quite used it.

# Allow this to fail, e.g. if running locally.
gcloud secrets versions access 1 --secret="storage-upload-key" > gcloud-key.json
gcloud auth activate-service-account --key-file gcloud-key.json

set -euo pipefail

export AIRTABLE_KEY=$(gcloud secrets versions access 1 --secret="airtable-key")

echo "Running forever..."
while true
do
	echo "Syncing..."
	gcloud auth activate-service-account --key-file gcloud-key.json
	./sync.sh
	sleep 60
done

