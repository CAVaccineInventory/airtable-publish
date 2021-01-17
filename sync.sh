#!/bin/bash
set -euf -o pipefail

mkdir -p airtable/safe
airtable-export --json airtable appy2N9zQSnFRPcN8 Locations --key $AIRTABLE_KEY

# Strip some fields.
cat airtable/Locations.json | jq 'del(.[]."Phone number")' | jq 'del(.[]."Add report")' | jq 'del(.[]."Add report link w/ phone number")' > airtable/safe/Locations.json

gsutil -h "Cache-Control:public, max-age=300" cp ./airtable/safe/Locations.json gs://cavaccineinventory-sitedata/airtable-sync/Locations.json

