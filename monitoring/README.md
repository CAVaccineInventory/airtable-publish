# Monitoring

"Making sure the pipeline is working"

## General Shape

PagerDuty (and Stackdriver) call Probes hosted on GCF which return success or
failure.

## Alerting

### PagerDuty

Checks every minute, and pages Vallery if the probe fails on production.

### Google Cloud Monitoring (Stackdriver)

Checks every minute, and posts to #operations if the probe fails in either staging or production.

[stackdriver-to-discord](https://github.com/Courtsite/stackdriver-to-discord) is deployed as a Cloud Function, and receives a [webhook](https://console.cloud.google.com/monitoring/alerting/notifications?project=cavaccineinventory#_0rif_slack-add-button:~:text=Webhooks,ADD%20NEW) from Stackdriver.

## Wishlist

* Size-delta (or absolute threshold) checks.
* Checking for a canary key.
* Rough data validation.  (Is it the right shape?)

## Why not something else?

No good reason.

Prometheus or similar would be nice, but then there's the overhead of
keeping the jobs running, and we don't want to spin up Kubernetes just
for that.
