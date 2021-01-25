# Intro

This directory contains python scripts to run in Google Cloud Functions for monitoring.

XXX more docs.


# Deployment to Google Cloud Functions

```
XXX WRITE ME
```


# Local development

```
docker build -t test-script . && \
docker run -it --rm \
 -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/key.json \
 -v $GOOGLE_APPLICATION_CREDENTIALS:/tmp/key.json:ro \
 test-script
```
