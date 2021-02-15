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
 - **Expected latency**: **130 seconds (2:10 min)**, with 30s due to
   having finished the pipeline that long ago, 40s for being at p75
   for how long before that it was published, and 60s for half of the
   browser cache.
 - **Maximum latency**: **300 seconds (5 min)**, with 60s from just
   missing the current publish, the previous publish having just
   beaten the 120s timeout, and all of that having filled the browser
   cache 120s ago.
 - **Minimal possible latency**: **30 seconds** from the pipeline just
   having completed in minimum observed time, and no browser cache.

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

2. Announce a push in #operations, and get a :thumbsup: from someone.

3. Run `scripts/deploy.sh`

4. Monitor production; same checks as in staging, above:
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

### Local-only

If you want to run locally without testing uploads, run
`./scripts/once.sh` and follow the instructions to set up your
configuration file with your Airtable key.  Output will be written to
the `local/` directory.


### Google Cloud testing

If you also want to test file uploads or metrics, you'll need to set
up your own Google Cloud project, monitoring workspace, and service
account:

1. Create a personal [project][projects] in Google Cloud.
2. Create a [monitoring workspace][workspaces] in that project.  If
   you have a fresh project, this should just be a matter of clicking
   on [Monitoring][monitoring].
3. Create a Google Cloud [service account][service-account] in there;
   name it something like `vaccinateca-testing`
4. [Grant][role-grants] the "Monitoring Metric Writer" and "Storage
   Object Admin" roles.
5. Download the key:

   ```
   gcloud iam service-accounts keys create testing-key.json --iam-account example@example.iam.gserviceaccount.com
   ```

6. Choose a unique bucket name to use (e.g. `alexmv-testing`)

7. Run `./scripts/pipeline.sh`, which will create a template `.env`
   file; edit it with your editor of choice, filling in the variables
   by following the instructions in it.

[projects]: https://cloud.google.com/resource-manager/docs/creating-managing-projects#creating_a_project
[workspaces]: https://cloud.google.com/monitoring/workspaces
[monitoring]: https://console.cloud.google.com/monitoring
[service-account]: https://cloud.google.com/iam/docs/creating-managing-service-accounts#creating
[role-grants]: https://cloud.google.com/iam/docs/granting-changing-revoking-access#granting-console

After that setup, you should be able to:
```
# To listen on port 8080
./scripts/pipeline.sh -bucket your-bucket-name-here

# In another terminal: curl -X POST http://localhost:8080/publish


# To publish once and exit
./scripts/once.sh -bucket your-bucket-name-here
```

### Adding a new resource type

1. Add a new function to `pipeline/pkg/airtable/tables.go` which calls
   `getTable` with the name of the table, as found in Airtable.

2. Determine the latest endpoint version, in
   `pipeline/pkg/endpoints/all.go`; since adding a new resource is
   backwards-compatible, we will not be increasing it.

3. Add a new package under `pipeline/pkg/endpoints/`; it should have a
   function named after the version (e.g. `V1`) which:

   1. Starts a beeline span
   2. Calls the function added in step 1, checking its err response
   3. Filters/modifies the columns using `filter.Transform`
   4. Returns the result.

4. Insert that function into `EndpointMap` in
   `pipeline/pkg/endpoints/all.go` under the latest version; the key
   should be the base filename the results are serialized as, the
   value should be the function you just wrote.

## Testing

```
go test -v ./...
```
