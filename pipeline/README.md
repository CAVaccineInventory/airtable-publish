# airtable-export

## How This Works

`airtable-export` is a worker that fetches data from Airtable upon a
POST to `/publish`, runs it through a sanitization pass (including but
not limited to removing superfluous or sensitive keys), then uploads
the results.

It fetches data using [the `airtable-export` Python
package](https://github.com/simonw/airtable-export)

It exposes current health status at `:8080`.  This can be used to
programmatically confirm if the most recent publish iteration
succeeded.

This tool exists because Airtable's API isn't feasible/safe to expose to client
Javascript, and Airtable has harsh rate limits.

## Production

The service runs on [Google Cloud
Run](https://console.cloud.google.com/run), and writes to multiple
Google Cloud Storage buckets:
 - https://console.cloud.google.com/storage/browser/cavaccineinventory-sitedata
 - https://console.cloud.google.com/storage/browser/vaccinateca-api
 - https://console.cloud.google.com/storage/browser/vaccinateca-api-staging
 
The former is served directly from
https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json,
the latter two are served through a
[CDN](https://console.cloud.google.com/net-services/cdn/details/api-vaccinateca-com)
and
[loadbalancer](https://console.cloud.google.com/net-services/loadbalancing/details/http/api-vaccinateca-com)
that serve https://api.vaccinateca.com/ and
https://staging-api.vaccinateca.com/

It is run every minute by hitting `/publish` with a [Cloud
Scheduler](https://console.cloud.google.com/cloudscheduler).


### Latencies

The job runs every minute, and takes on average 45s to complete; it
uploads with caching headers that instruct clients to cache the data
for 2 minutes.

For a JSON request that a browser requests:
 - **Expected latency**: **140 seconds (2:20 min)**, with 30s due to
   having finished the pipeline that long ago, 50s for being at p99
   for how long before that it was published, and 60s for half of the
   browser cache.
 - **Maximum latency**: **300 seconds (5 min)**, with 60s from just
   missing the current publish, the previous publish having just
   beaten the 120s timeout, and all of that having filled the browser
   cache 120s ago.
 - **Minimal possible latency**: **45 seconds** from the pipeline just having
   completed, and no browser cache.

Note that the _maximum latency_ above is _if it is still publishing
data._  If the pipeline hangs for longer than the 2 minute timeout,
then it will not make forward progress, and no updates will be
written.  Once the file is 10 minutes stale, the monitoring will page
(see `monitoring/freshcf`).

### Deploys

**The `main` branch auto-deploys to staging, the `prod` branch to
production.**

Deploys take ~2 minutes to complete, and are controlled through
[Google Cloud
Build](https://console.cloud.google.com/cloud-build/triggers).

To deploy staging to production:

1. Verify that staging is happy (ordered below from broad to fiddly):
   - [Staging monitoring](https://freshcf-staging-patvwfu2ya-uw.a.run.app/)
     should be `OK`
   - [Dashboard](https://console.cloud.google.com/monitoring/dashboards/builder/75b273d3-6724-48d0-8dad-0922f6207f79)
     should most publishes taking <60s, no failures.
   - [Logs](https://console.cloud.google.com/run/detail/us-west1/airtable-export-staging/logs)
     should show no warnings or failures.
   - [Service URL itself](https://airtable-export-staging-patvwfu2ya-uw.a.run.app/healthcheck)

2. [Create a pull request from `main` into `prod`](https://github.com/CAVaccineInventory/airtable-export/compare/prod...main?quick_pull=1&title=[DEPLOY]+%28summarize%20here%29)
   - Describe the key changes in the summary, and any notes in the body.

3. Get that pull request reviewed and accepted.

4. Merge the pull request **as a merge**.  Merging it as a _rebase_
   will cause divergent history between `main` and `prod` which
   requires a force-push to fix.

5. Monitor production; same checks as in staging, above:
   - [Monitoring](https://freshcf-prod-patvwfu2ya-uw.a.run.app/)
     will page if it is not `OK`
   - [Dashboard](https://console.cloud.google.com/monitoring/dashboards/builder/75b273d3-6724-48d0-8dad-0922f6207f79)
   - [Logs](https://console.cloud.google.com/run/detail/us-west1/airtable-export-prod/logs)
   - [Service URL itself](https://airtable-export-prod-patvwfu2ya-uw.a.run.app/healthcheck)

## Secrets

In production, these are fetched automatically from Google Cloud's Secrets Manager;
 - [airtable-key](https://console.cloud.google.com/security/secret-manager/secret/airtable-key)
 - [honeycomb-key](https://console.cloud.google.com/security/secret-manager/secret/honeycomb-key)
 - [storage-upload-key](https://console.cloud.google.com/security/secret-manager/secret/storage-upload-key)

## Development

In development:
 - You can look up your Airtable API key at https://airtable.com/account
 - You can look up the testing Honeycomb API key at https://ui.honeycomb.io/teams/vaccinateca
 - Set up your own service account for testing:
    1. Create your own test workspace, if you do no have one already
    2. Create a Google Cloud service account in there
    3. Grant it "Monitoring Metric Writer" and "Storage Object Admin" rights.
    4. Download the key:

       ```
       gcloud iam service-accounts keys create testing-key.json --iam-account example@example.iam.gserviceaccount.com
       ```

 - Choose a unique bucket name to use
 - Build and run in docker:

```
docker build -t airtable-export .

docker run \
  -e AIRTABLE_KEY=<key> \
  -e TESTING_BUCKET=<bucketname> \
  -v "$(pwd)/testing-key.json:/testing-key.json" \
  -p 8080:8080
  --rm -it airtable-export /entrypoint.sh once
```

## Testing

```
go test -v ./...
```
