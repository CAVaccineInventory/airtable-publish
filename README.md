This repo fetches data from Airtable, and publishes it to a storage bucket, in JSON format.
This tool exists because Airtable's API isn't feasible/safe to expose to client Javascript.

# Setup

* Build sanitize binary.
* Create a cron/contab entry for sync.sh, with `$AIRTABLE_KEY` set.
