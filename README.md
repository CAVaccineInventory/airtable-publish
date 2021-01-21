This repo fetches data from Airtable, and publishes it to a storage bucket, in JSON format.
This tool exists because Airtable's API isn't feasible/safe to expose to client Javascript,
and Airtable has harsh rate limits.

# Invocation

Environment variables:

* AIRTABLE_KEY: airtable API key
* BUCKET_PATH: fully-qualified Google Cloud Storage bucket path (e.g. `gs://bucket/dir1/dir2`).

Secrets:

* /gcloud-key.json: a Google Cloud service account key, with write access to the storage bucket

Example docker invokation:
`docker run -e AIRTABLE_KEY=<key> -e BUCKET_PATH=gs://gs://cavaccineinventory-sitedata/<directory> -v <gcloud storage key>:/gcloud-key.json -it <image>`

# Old Setup

* Build sanitize binary.
* Create a cron/contab entry for sync.sh, with `$AIRTABLE_KEY` set.
