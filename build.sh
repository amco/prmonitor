#!/bin/bash

set -e
gofmt -l -w .
test -z "$(gofmt -s -d .)"

# fetch vendor dep?
godep update ./...


# will it build?
godep go install github.com/brentdrich/prmonitor/cmd/prmonitor

# will it test?
godep go test

# will it lint?
deadcode ./..
golint github.com/brentdrich/prmonitor
errcheck -ignore '[rR]ead|[wW]rite|[cC]lose|[sS]top' github.com/brentdrich/prmonitor
interfacer github.com/brentdrich/prmonitor
unconvert github.com/brentdrich/prmonitor