# How This Works
`airtable-export` is a worker that periodically fetches data from Airtable,
runs it through a sanitization pass (including but not limited to removing superfluous or sensitive keys),
then uploads the results.

It exposes current health status at `:8080`.
THis can be used to programmatically confirm if the most recent publish iteration succeeded.

This tool exists because Airtable's API isn't feasible/safe to
expose to client Javascript, and Airtable has harsh rate limits.

# Production (Actually Staging)

Runs at: https://console.cloud.google.com/run/detail/us-west1/airtable-exporter/metrics?project=cavaccineinventory
Pushes data to: [gs://cavaccineinventory](https://console.cloud.google.com/storage/browser/cavaccineinventory-sitedata?project=cavaccineinventory&pageState=(%22StorageObjectListTable%22:(%22f%22:%22%255B%255D%22))&prefix=&forceOnObjectsSortingFiltering=false)

# Invocation

Environment variables:

- AIRTABLE_KEY: airtable API key
- BUCKET_PATH: fully-qualified Google Cloud Storage bucket path (e.g.
  `gs://bucket/dir1/dir2`).

Secrets:

In production, this is fetched automatically.
In development, you should mount the file if you need to upload as part of your QA cycle.

- /gcloud-key.json: a Google Cloud service account key, with write access to the
  storage bucket

## Development

Example docker invokation:
`docker run -e AIRTABLE_KEY=<key> -e BUCKET_PATH=gs://gs://cavaccineinventory-sitedata/<directory> -v <gcloud storage key>:/gcloud-key.json -it <image>`

# Old Setup

- Build sanitize binary.
- Create a cron/contab entry for sync.sh, with `$AIRTABLE_KEY` set.

# Testing

- Requires the _gron_ tool: https://github.com/tomnomnom/gron

# Fields in use by the 'site' repo

Run `pipenv run ./get_required_fields_for_site.py ../../site/assets/js/data.js`
inside the `sanitize` directory. Adjust the path to `data.js` according to where
you have the file.
