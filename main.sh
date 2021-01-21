#!/bin/bash
set -euo pipefail

echo "Running forever..."
while true
do
	echo "Syncing..."
	gcloud auth activate-service-account --key-file gcloud-key.json
	./sync.sh
	sleep 60
done

