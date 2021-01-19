#!/bin/bash
set -euf -o pipefail

OUTDIR=airtable

# Fetch data.
mkdir -p $OUTDIR/safe
/usr/local/bin/airtable-export --json $OUTDIR appy2N9zQSnFRPcN8 Locations --key $AIRTABLE_KEY

# Clean up data.
# This includes a sanitization pass, and removing data that we don't want to widely publish.
# For example, phone numbers are easy to present without proper context, and DDOS a location.
# In other cases, some fields contain privileged links.

# PLEASE ASK VALLERY, MANISH, OR ANOTHER DATA EXPERT BEFORE CHANGING THIS.

# Locations.json is a slightly reduced version of the main dataset.
./sanitize/sanitize $OUTDIR/Locations.json | \
  jq 'del(.[]."Add report")' | \
  jq 'del(.[]."Add report link w/ phone number")' | \
  jq 'del(.[]."airtable_createdTime")' | \
  jq 'del(.[]."Internal notes")' | \
  jq 'del(.[]."Phone number")' | \
  jq -c \
  > $OUTDIR/safe/Locations.json

# Upload data.
gsutil -h "Cache-Control:public, max-age=300" cp -Z $OUTDIR/safe/Locations.json gs://cavaccineinventory-sitedata/airtable-sync/Locations.json
