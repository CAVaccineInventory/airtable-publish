# Monitoring

"Making sure the pipeline is working"

## General Shape

PagerDuty calls Probes hosted on GCF which return success or failure.

## Wishlist

* Size-delta (or absolute threshold) checks.
* Checking for a canary key.
* Rough data validation.  (Is it the right shape?)

## Why not something else?

No good reason.

Prometheus or similar would be nice, but then there's the overhead of
keeping the jobs running, and we don't want to spin up Kubernetes just
for that.
