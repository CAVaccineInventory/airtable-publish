# freshcf

A worker for testing freshness of Locations.json; runs as a Cloud Run
deploy for each of
[staging](https://console.cloud.google.com/run/detail/us-west1/freshcf-staging)
and
[prod](https://console.cloud.google.com/run/detail/us-west1/freshcf-prod).

Returns 200 if everything is ok; returns 500 and a text description
otherwise, suitable for using with a simple prober service that can
look at response code.

 - The `/` URL returns a status code
 - The `/json` URL returns metadata about all of the endpoints it is monitoring, as JSON.
 - The `/push` URL, when POST'd to every minute by Cloud Scheduler,
   pushes metrics about the published JSON to Stackdriver.

## Deployment

**The `main` branch auto-deploys to staging, the `prod-monitoring`
(*not* `prod`!) branch to production.**

Deploys take ~2 minutes to complete, and are controlled through
[Google Cloud
Build](https://console.cloud.google.com/cloud-build/triggers).

To deploy staging to production:

1. [Create a pull request from `main` into `prod-monitoring`](https://github.com/CAVaccineInventory/airtable-export/compare/prod-monitoring...main?quick_pull=1&title=[DEPLOY]+%28summarize%20here%29)
   - Describe the key changes in the summary, and any notes in the body.

2. Get that pull request reviewed and accepted.

3. Merge the pull request **as a merge**.  Merging it as a _rebase_
   will cause divergent history between `main` and `prod` which
   requires a force-push to fix.


## Local Development

``` shell
go run cmd/server/main.go
curl http://localhost:8080/
```

## Potential Future Development

* Add `threshold` as a query parameter.
