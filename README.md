# OneSignal Cleaner

Deletes outdated players

# Docker

### Docker Hub

```shell
docker run --rm mingalevme/onesignal-cleaner --app-id "your-app-id" --rest-api-key "your-app-rest-api-key" --ttl $(( 86400*30*12 )) --debug
```

or via environment vars:

```shell
docker run --rm \
  -e "ONESIGNAL_CLEANER_APP_ID=app-id" \
  -e "ONESIGNAL_CLEANER_REST_API_KEY=rest-api-key" \
  -e "ONESIGNAL_CLEANER_TTL=$(( 86400*30*12 ))" \
  -e "ONESIGNAL_CLEANER_DEBUG=1" \
  mingalevme/onesignal-cleaner
```

## Build

```shell
docker build --target app -t onesignal-cleaner .
```

# Run via the code

```shell
go run onesignal-cleaner --app-id "your-app-id" --rest-api-key "your-app-rest-api-key" --ttl $(( 86400*30*12 ))
```

# Develop

## Testing

```shell
go test -v -cover -tags testing ./...
```
