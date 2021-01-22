#!/bin/bash
set -euf -o pipefail

OUTDIR=`dirname $0`/../airtable
TEST_DATA_DIR=`dirname $0`/test_data

/usr/local/bin/airtable-export --json $OUTDIR appy2N9zQSnFRPcN8 Locations --key $AIRTABLE_KEY

# Just take the first 100 entries of Locations.json
gron $OUTDIR/Locations.json | rg 'json\[[\d]{1}\]' | gron -ungron > ${TEST_DATA_DIR}/reduced_locations.json
