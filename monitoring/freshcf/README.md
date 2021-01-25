# freshcf

A cloud function for testing freshness of Locations.json.

Returns 200 if everything is ok.  Returns 500 otherwise.

Suitable for using with a simple prober service that can look at response code.

## Deployment

``` shell
$ gcloud functions deploy \
  freshLocations \
  --project cavaccineinventory \
  --entry-point CheckFreshness \
  --runtime go113 \
  --trigger-http \
  --allow-unauthenticated \
  --source=.
```

Deployment can take up to two minutes, as under the hood it builds a new container and does other magic.

## Local Development

``` shell
go run cmd/main.go
curl http://localhost:8080/
```

## Functions Framework

References:

* https://cloud.google.com/functions/docs/functions-framework
* https://github.com/GoogleCloudPlatform/functions-framework-go

## Potential Future Development

* Add `threshold` as a query parameter.
* Support probing multiple URLs.  This could be some sort of query parameter
  (maybe to a lookup table) or by using multiple entry points and independent cloud functions.
* Reduce the amount of memory configured for the function.  Microoptimization,
  probably would save some cents.
  