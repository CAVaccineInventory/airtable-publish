# airtable-export

## How This Works

`airtable-export` is a worker that periodically fetches data from
Airtable, runs it through a sanitization pass (including but not
limited to removing superfluous or sensitive keys), then uploads the
results.

It fetches data using [the `airtable-export` Python
package](https://github.com/simonw/airtable-export)

It exposes current health status at `:8080`.
This can be used to programmatically confirm if the most recent publish
iteration succeeded.

This tool exists because Airtable's API isn't feasible/safe to expose to client
Javascript, and Airtable has harsh rate limits.

## Production

**The `main` branch auto-deploys to staging, the `prod` branch to production.**

Deploys take ~2 minutes to complete, and published files are cached
for a few minutes.

Auto-deployment configured through Cloud Build.
[trigger](https://console.cloud.google.com/cloud-build/triggers/edit/2a8c6015-8b1d-4073-815f-f35edd1a3b1a?project=cavaccineinventory)

* Runs on [Google Cloud Run](https://console.cloud.google.com/run/detail/us-west1/airtable-export-prod/metrics?authuser=1&organizationId=0&project=cavaccineinventory&supportedpurview=project)

(serverless thing that handles automatic container build and deploy).

* Pushes data to: [gs://cavaccineinventory-sitedata](https://console.cloud.google.com/storage/browser/cavaccineinventory-sitedata?project=cavaccineinventory&pageState=(%22StorageObjectListTable%22:(%22f%22:%22%255B%255D%22))&prefix=&forceOnObjectsSortingFiltering=false)

Google Cloud Run throttles the app to a crawl when not handling a request.
As a _terrible_ workaround, the health check endpoint sleeps for >= 1 minute.
We run a [magic cloud cronjob](https://console.cloud.google.com/cloudscheduler?project=cavaccineinventory)
which calls that status/healthcheck endpoint minutely as a keep-warm.
We know it's messy - improve it if and only if it's worth the time and
opportunity cost.

## Invocation

### Environment variables

* AIRTABLE_KEY: airtable API key

### Secrets

In production, these are fetched automatically from Google Cloud's Secrets Manager;
 - [airtable-key](https://console.cloud.google.com/security/secret-manager/secret/airtable-key/versions?project=cavaccineinventory)
 - [storage-upload-key](https://console.cloud.google.com/security/secret-manager/secret/storage-upload-key/versions?project=cavaccineinventory)

In development:
 - You can look up your Airtable API key at https://airtable.com/account
 - Create Google Cloud service account key, with write access to the
   storage bucket; you should mount the file into the Docker image if
   you need to upload as part of your QA cycle.


## Development

Example docker invocation:

```
docker build -t airtable-export

docker run \
  -e AIRTABLE_KEY=<key> \
  -v <gcloud storage key>:/gcloud-key.json \
  --rm -it airtable-export
```

## Testing

```
go test -v ./...
```

## Fields in use by the 'site' repo

Run `pipenv run ./get_required_fields_for_site.py ../../site/assets/js/data.js`

inside the `sanitize` directory. Adjust the path to `data.js` according to where
you have the file.
