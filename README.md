# notifier-cli

### Sample program to handle arbitrarily large list messages to a specific url, using semaphoes to control the number of spawned routines where the url call is made
http POST requests to a configured url.

Data is read from STDIN and the rate of sending can be controlled via
`--i or --interval eg --i=5s` parameter
### Usage:

The folder contains a sample `data.txt` file for testing.

Replace `--url` value with a valid url.

1. Run directly
```
go run main.go --url=https://webhook.site/f30d570f-20ac-4211-a389-2f3696e1fa45 < data.txt --i=3s
```

2. Build binary and run
```
go build

./notifier-cli --url=https://webhook.site/f30d570f-20ac-4211-a389-2f3696e1fa45 < data.txt --i=3s
```

### Tests:
tests are available for the notifier library/package
```
go test ./...
```
