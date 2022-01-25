# OneSignal Cleaner

Deletes outdated players

# Docker

## Build

```shell
docker build --target app -t onesignal-cleaner .
```

### Run

```shell
docker run --rm \
  -e "ONESIGNAL_CLEANER_APP_ID=app-id" \
  -e "ONESIGNAL_CLEANER_REST_API_KEY=rest-api-key" \
  -e "ONESIGNAL_CLEANER_TTL=15552000" \
  -e "ONESIGNAL_CLEANER_DEBUG=1" \
  onesignal-cleaner

# Run via the code

```shell
go run onesignal-cleaner --app-id "your-app-id" --rest-api-key "your-app-rest-api-key" --ttl $(( 86400*30*12 ))
```
