#!/bin/bash
set -euo pipefail

# BUCKET_PATH
OUTDIR=airtable

# Fetch data.
mkdir -p $OUTDIR/safe
airtable-export --json $OUTDIR appy2N9zQSnFRPcN8 Locations --key $AIRTABLE_KEY

# Clean up data.
# This includes a sanitization pass, and removing data that we don't want to widely publish.
# For example, phone numbers are easy to present without proper context, and DDOS a location.
# In other cases, some fields contain privileged links.

# PLEASE ASK VALLERY, MANISH, OR ANOTHER DATA EXPERT BEFORE CHANGING THIS.

# Locations.json is a slightly reduced version of the main dataset.
./sanitize/sanitize $OUTDIR/Locations.json
./sanitize/sanitize $OUTDIR/Locations.json | \
  jq -c \
  > $OUTDIR/safe/Locations.json

# Upload data.
gsutil -h "Cache-Control:public, max-age=300" cp -Z $OUTDIR/safe/Locations.json $BUCKET_PATH/Locations.json
