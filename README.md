# notifier-cli
### (exercise only, non-prod)

This project file contains a notifier package for sending
http POST requests to a configured url.

Data is read from STDIN and the rate of sending can be controlled via
`--i or --interval eg --i=5s` parameter
###Usage:

The folder contains a sample `data.txt` file for testing

Run directly
```
go run main.go --url=https://webhook.site/f30d570f-20ac-4211-a389-2f3696e1fa45 < data.txt
```

Build a binary
```
go build

./notifier-cli --url=https://webhook.site/f30d570f-20ac-4211-a389-2f3696e1fa45 < data.txt
```

###Tests:
tests are available for the notifier library/package
```
go test ./...
```