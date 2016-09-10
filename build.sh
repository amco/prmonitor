#!/bin/bash

set -e
gofmt -l -w .
test -z "$(gofmt -s -d .)"

# will it build?
godep go install github.com/brentdrich/prmonitor/cmd/prmonitor

# will it test?
go test

# will it lint?
deadcode ./cmd/prmonitor
golint github.com/brentdrich/prmonitor/cmd/prmonitor
errcheck github.com/brentdrich/prmonitor/cmd/prmonitor
interfacer github.com/brentdrich/prmonitor/cmd/prmonitor
unconvert github.com/brentdrich/prmonitor/cmd/prmonitor