# Intro

This directory contains python scripts to run in Google Cloud Functions for monitoring.

Currenly the main use is exporting metrics about the output files from airtable-exporter into Google Cloud Monitoring (aka Stackdriver).

This could at some point get ported to Go and integrated better with the other monitoring code and the rest of the repo. This being in python is a matter of expediency, not preference.

# Deployment to Google Cloud Functions

This code runs in Google Cloud Functions. It is triggered every minute by an action in Google Cloud Scheduler.

## IAM Setup

To post to monitoring, this need to run with a Service Account that has Monitoring Editor. The `monitoring` service account was created by hand for this purpose.

## Function deployment

Do this every time the code is updated. Note that you must have the 'Cloud Functions Admin' permissions, otherwise you may get an error about not being able to set IAM permissions on the project and the function will not be publically accessable.

There are currently 4 monitored URLs, 2 for prod and 2 for staging. Each URL gets its own function. Perhaps wasteful, but it makes it easy to reconfigure as things change.

Here are the invocations for each URL. The only differences are the TARGET_* variables.

```
gcloud functions deploy \
  gcsfileStatsLocationsProd \
  --project cavaccineinventory \
  --entry-point main \
  --runtime python38 \
  --trigger-http \
  --allow-unauthenticated \
  --service-account monitoring@cavaccineinventory.iam.gserviceaccount.com \
  --set-env-vars=TARGET_URL=https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json \
  --set-env-vars=TARGET_ENV=prod \
  --set-env-vars=TARGET_NAME=locations \
  --source=.

gcloud functions deploy \
  gcsfileStatsCountiesProd \
  --project cavaccineinventory \
  --entry-point main \
  --runtime python38 \
  --trigger-http \
  --allow-unauthenticated \
  --service-account monitoring@cavaccineinventory.iam.gserviceaccount.com \
  --set-env-vars=TARGET_URL=https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Counties.json \
  --set-env-vars=TARGET_ENV=prod \
  --set-env-vars=TARGET_NAME=counties \
  --source=.


gcloud functions deploy \
  gcsfileStatsLocationsStaging \
  --project cavaccineinventory \
  --entry-point main \
  --runtime python38 \
  --trigger-http \
  --allow-unauthenticated \
  --service-account monitoring@cavaccineinventory.iam.gserviceaccount.com \
  --set-env-vars=TARGET_URL=https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync-staging/Locations.json \
  --set-env-vars=TARGET_ENV=staging \
  --set-env-vars=TARGET_NAME=locations \
  --source=.

gcloud functions deploy \
  gcsfileStatsCountiesStaging \
  --project cavaccineinventory \
  --entry-point main \
  --runtime python38 \
  --trigger-http \
  --allow-unauthenticated \
  --service-account monitoring@cavaccineinventory.iam.gserviceaccount.com \
  --set-env-vars=TARGET_URL=https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync-staging/Counties.json \
  --set-env-vars=TARGET_ENV=staging \
  --set-env-vars=TARGET_NAME=counties \
  --source=.


```


## Scheduler deployment

This is what causes the functions to run once every minute to push stats to monitoring.

This only needs to be done once, after the functions are set up.

```
gcloud scheduler jobs create http gcsfileStatsLocationsProd --schedule='* * * * *' --uri=https://us-central1-cavaccineinventory.cloudfunctions.net/gcsfileStatsLocationsProd

gcloud scheduler jobs create http gcsfileStatsCountiesProd --schedule='* * * * *' --uri=https://us-central1-cavaccineinventory.cloudfunctions.net/gcsfileStatsCountiesProd

gcloud scheduler jobs create http gcsfileStatsLocationsStaging --schedule='* * * * *' --uri=https://us-central1-cavaccineinventory.cloudfunctions.net/gcsfileStatsLocationsStaging

gcloud scheduler jobs create http gcsfileStatsCountiesStaging --schedule='* * * * *' --uri=https://us-central1-cavaccineinventory.cloudfunctions.net/gcsfileStatsCountiesStaging

```



# Local development

To run this manually in development, you need a private key file from GCP to authorize posting to Cloud Monitoring.

Here is an example invocation to use a key to publish to a separate account from your laptop.

```

export GOOGLE_APPLICATION_CREDENTIALS=`pwd`/debug-key.json

docker build -t test-script . && \
docker run -it --rm \
 -e TARGET_URL=https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Counties.json \
 -e TARGET_ENV=prod \
 -e TARGET_NAME=counties \
 -e PROJECT_NAME=personal-project-name \
 -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/key.json \
 -v $GOOGLE_APPLICATION_CREDENTIALS:/tmp/key.json:ro \
 test-script
```
