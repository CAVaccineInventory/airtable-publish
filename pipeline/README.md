# airtable-export

## How This Works

`airtable-export` is a worker that fetches data from Airtable upon a
POST to `/publish`, runs it through a sanitization pass (including but
not limited to removing superfluous or sensitive keys), then uploads
the results.

It fetches data using [the `airtable-export` Python
package](https://github.com/simonw/airtable-export)

It exposes current health status at `:8080`.
This can be used to programmatically confirm if the most recent publish
iteration succeeded.

This tool exists because Airtable's API isn't feasible/safe to expose to client
Javascript, and Airtable has harsh rate limits.

## Production

The service runs on [Google Cloud
Run](https://console.cloud.google.com/run?project=cavaccineinventory),
and writes to [Google Cloud
Storage](https://console.cloud.google.com/storage/browser/cavaccineinventory-sitedata).

### Deploys

**The `main` branch auto-deploys to staging, the `prod` branch to
production.**

Deploys take ~2 minutes to complete, and are controlled through
[Google Cloud
Build](https://console.cloud.google.com/cloud-build/triggers?project=cavaccineinventory).

To deploy staging to production:

1. Verify that staging is happy (ordered below from broad to fiddly):
   - [Staging monitoring](https://us-central1-cavaccineinventory.cloudfunctions.net/freshLocationsStaging)
     should be `OK`
   - [Dashboard](https://console.cloud.google.com/monitoring/dashboards/builder/75b273d3-6724-48d0-8dad-0922f6207f79?project=cavaccineinventory)
     should most publishes taking <60s, no failures, and 0.02 success/sec.
   - [Logs](https://console.cloud.google.com/run/detail/us-west1/airtable-export-staging/logs?project=cavaccineinventory)
     should show no warnings or failures.
   - [Service URL itself](https://airtable-export-staging-patvwfu2ya-uw.a.run.app/healthcheck)

2. [Create a pull request from `main` into `prod`](https://github.com/CAVaccineInventory/airtable-export/compare/prod...main?quick_pull=1&title=[DEPLOY]+%28summarize%20here%29)
   - Describe the key changes in the summary, and any notes in the body.

3. Get that pull request reviewed and accepted.

4. Merge the pull request _as a rebase_.  This should be the only option.

5. Monitor production; same checks as in staging, above:
   - [Monitoring](https://us-central1-cavaccineinventory.cloudfunctions.net/freshLocations)
     will page if it is not `OK`
   - [Dashboard](https://console.cloud.google.com/monitoring/dashboards/builder/75b273d3-6724-48d0-8dad-0922f6207f79?project=cavaccineinventory)
   - [Logs](https://console.cloud.google.com/run/detail/us-west1/airtable-export-prod/logs?project=cavaccineinventory)
   - [Service URL itself](https://airtable-export-prod-patvwfu2ya-uw.a.run.app/healthcheck)

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
