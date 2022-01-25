# OneSignal Cleaner

Deletes outdated players

# Run via the code

```shell
go run onesignal-cleaner --app-id "your-app-id" --rest-api-key "your-app-rest-api-key" --ttl $(( 86400*30*12 ))
```

# Development hints

## Update gz_csv_reader_test_data.csv.gz

```shell
gzip -c gz_csv_reader_test_data.csv > gz_csv_reader_test_data.csv.gz
```
