# airtable-export

## How This Works

`airtable-export` is a worker that periodically fetches data from Airtable, runs
it through a sanitization pass (including but not limited to removing
superfluous or sensitive keys), then uploads the results.

It exposes current health status at `:8080`.
This can be used to programmatically confirm if the most recent publish
iteration succeeded.

This tool exists because Airtable's API isn't feasible/safe to expose to client
Javascript, and Airtable has harsh rate limits.

## Production

**The main branch auto-deploys to production.**

Production deploys take a few minutes to complete, and published files are
cached for a few minutes.

Auto-deployment configured through Cloud Build.
[trigger](https://console.cloud.google.com/cloud-build/triggers/edit/2a8c6015-8b1d-4073-815f-f35edd1a3b1a?project=cavaccineinventory)

**Not sure if this is working?**
Check the headers on the
[locations](https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json)
file (via curl, browser tools, whatever).
`Last-Modified` should be minutes old.

For example:
```
$ curl -sI https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json | grep last-modified; echo -n "          now: "; date -Ru
last-modified: Sun, 24 Jan 2021 05:34:47 GMT
          now: Sun, 24 Jan 2021 05:57:58 +0000
```

* Runs on [Google Cloud Run](https://console.cloud.google.com/run/detail/us-west1/airtable-export-prod/metrics?authuser=1&organizationId=0&project=cavaccineinventory&supportedpurview=project)

(serverless thing that handles automatic container build and deploy).

* Pushes data to: [gs://cavaccineinventory](https://console.cloud.google.com/storage/browser/cavaccineinventory-sitedata?project=cavaccineinventory&pageState=(%22StorageObjectListTable%22:(%22f%22:%22%255B%255D%22))&prefix=&forceOnObjectsSortingFiltering=false)

Google Cloud Run throttles the app to a crawl when not handling a request.
As a _terrible_ workaround, the health check endpoint sleeps for >= 1 minute.
We run a [magic cloud cronjob](https://console.cloud.google.com/cloudscheduler?project=cavaccineinventory)
which calls that status/healthcheck endpoint minutely as a keep-warm.
We know it's messy - improve it if and only if it's worth the time and
opportunity cost.

## Invocation

### Environment variables

* AIRTABLE_KEY: airtable API key
* BUCKET_PATH: fully-qualified Google Cloud Storage bucket path (e.g.
  `gs://bucket/dir1/dir2` ).

### Secrets

In production, these are fetched automatically.
In development, you should mount the file if you need to upload as part of your QA cycle.

* /gcloud-key.json: a Google Cloud service account key, with write access to the
  storage bucket

## Development

Example docker invocation:

```
docker run \
  -e AIRTABLE_KEY=<key> \
  -e BUCKET_PATH=gs://cavaccineinventory-sitedata/<directory> \
  -v <gcloud storage key>:/gcloud-key.json \
  --rm -it <image>`
```

## Testing

```
go test -v ./...
```

## Fields in use by the 'site' repo

Run `pipenv run ./get_required_fields_for_site.py ../../site/assets/js/data.js`

inside the `sanitize` directory. Adjust the path to `data.js` according to where
you have the file.
