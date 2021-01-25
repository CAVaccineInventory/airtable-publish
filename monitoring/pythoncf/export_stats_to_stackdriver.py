#!/usr/bin/env python3

# Meant to be placed in google cloud functions. See README.md for deployment and testing instructions.

import os
import requests
from dateutil.parser import parse
import json
import time
import datetime
from google.cloud import monitoring_v3


# Which environment is this (eg, prod, staging).
TARGET_ENV = os.environ.get('TARGET_ENV', 'prod')
# Name of the file we're monitoring. Currently we only have Locations.json, but we can have more later.
TARGET_NAME = os.environ.get('TARGET_NAME', 'locations')
# URL to fetch.
TARGET_URL = os.environ.get(
    'TARGET_URL',
    'https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json')

# Name of the GCP project to submit metrics to.
PROJECT_NAME = os.environ.get('PROJECT_NAME', 'cavaccineinventory')

def main(request):

    # get the target URL
    target_response = requests.get(TARGET_URL)

    # compute age and length
    target_modified = target_response.headers['last-modified']
    target_age = (datetime.datetime.now(datetime.timezone.utc) -
                  parse(target_modified)).total_seconds()
    target_length = 0
    try:
        target_length = len(target_response.json())
    except:
        pass

    # keys to output
    output = {
        'last_modifided_age_seconds': target_age,
        'length_bytes': len(target_response.content),
        'length_json_items': target_length,
    }

    # write to stackdriver
    # https://cloud.google.com/monitoring/custom-metrics/creating-metrics
    client = monitoring_v3.MetricServiceClient()
    project_name = f"projects/{PROJECT_NAME}"

    for metric_name, metric_value in output.items():

        series = monitoring_v3.TimeSeries()
        series.metric.type = f'custom.googleapis.com/gcsfile/{metric_name}'
        series.resource.type = 'generic_node'
        series.resource.labels['project_id'] = PROJECT_NAME
        series.resource.labels['location'] = 'us-central'
        series.resource.labels['node_id'] = TARGET_NAME
        series.resource.labels['namespace'] = TARGET_ENV

        now = time.time()
        seconds = int(now)
        nanos = int((now - seconds) * 10 ** 9)
        interval = monitoring_v3.TimeInterval(
            {"end_time": {"seconds": seconds, "nanos": nanos}}
        )

        point = monitoring_v3.Point(
            {"interval": interval, "value": {
                "double_value": output[metric_name] }})

        series.points = [point]

        client.create_time_series(name=project_name, time_series=[series])


    return json.dumps(output, indent=2)

if __name__ == '__main__':
    print(main(None))
