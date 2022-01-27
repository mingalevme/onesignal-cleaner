# OneSignal Cleaner

Deletes outdated players

# Docker

### Docker Hub

```shell
docker run --rm mingalevme/onesignal-cleaner \
  --app-id "your-app-id" \
  --rest-api-key "your-app-rest-api-key" \
  --inactivity-threshold "$(( 86400*30*12 ))" \
  --concurrency 10
  --debug
```

or via environment vars:

```shell
docker run --rm \
  -e "ONESIGNAL_CLEANER_APP_ID=app-id" \
  -e "ONESIGNAL_CLEANER_REST_API_KEY=rest-api-key" \
  -e "ONESIGNAL_CLEANER_INACTIVITY_THRESHOLD=$(( 86400*30*12 ))" \
  -e "ONESIGNAL_CLEANER_CONCURRENCY=10" \
  -e "ONESIGNAL_CLEANER_DEBUG=1" \
  mingalevme/onesignal-cleaner
```

or 

```shell
(
  APP_ID="app-id"
  REST_API_KEY="rest-api-key"
  INACTIVITY_THRESHOLD="$((86400*365))"
  CONCURRENCY="10"
  NOW=`date +"%Y-%m-%dT%H:%M:%S%z"`
  docker run --rm mingalevme/onesignal-cleaner \
    --app-id "$APP_ID" \
    --rest-api-key "$REST_API_KEY" \
    --inactivity-threshold "$INACTIVITY_THRESHOLD"
    --concurrency "$CONCURRENCY" \
    | tee "./onesignal-cleaner-$NOW-$APP_ID.log"
)
```

## Build

```shell
docker build --target app -t onesignal-cleaner .
```

# Run via the code

```shell
go run ./... --app-id "your-app-id" --rest-api-key "your-app-rest-api-key" --inactivity-threshold $(( 86400*30*12 )) --concurrency 10 --debug
```

# Develop

## Testing

```shell
go test -v -cover -tags testing ./...
```
