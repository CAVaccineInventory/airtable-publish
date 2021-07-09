# airtable-export

This API exporter is deprecated following the introduction of [Vial](https://github.com/CAVaccineInventory/vial).

## Access
* Github committer team: [site-data-committers](https://github.com/orgs/CAVaccineInventory/teams/site-data-committers)
* Discord channels:
    * #data-publishing
    * #operations (infrastructure and production)
* Google Cloud access managers: @Vallery, @Robert S, @Alex V
    * Logs
    * Monitoring
    * Deployments
    * Storage buckets (for serving the "API")
* Airtable access
* Honeycomb access: [self-request](https://ui.honeycomb.io/join_team/vaccinateca)
* Pagerduty access managers: @Vallery


## How This Works

`airtable-export` is a worker that periodically fetches data from
Airtable, runs it through a sanitization pass (including but not
limited to removing superfluous or sensitive keys), then uploads the
results.  The data is written in multiple formats to multiple places
(e.g. https://api.vaccinateca.com/v1/ as well as the legacy
https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json).
A staging and a production deploy exist, which write to different
places.

`monitoring` is a black-box monitoring of the output of that, which is
used to page on failures.  One is deployed for each of prod and
staging; they each monitor all of the URLs which that deploy writes to.

## Layout

* `pipeline` contains the main pipeline code.
* `monitoring` contains directories containing monitoring related code
  * `monitoring/freshcf` is the freshness monitor and prober.

## Linting

We use [golangci](https://golangci-lint.run/)'s linter wrapper.

To run locally, run:

``` shell
docker run --rm -v $(pwd):/app -w /app \
    golangci/golangci-lint:v1.35.2 golangci-lint run \
    -E golint,goimports,misspell
```
