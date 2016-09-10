#!/bin/bash

set -e
gofmt -l -w .
test -z "$(gofmt -s -d .)"

# will it build?
godep go install github.com/brentdrich/prmonitor

# will it lint?
deadcode .
golint github.com/brentdrich/prmonitor
errcheck github.com/brentdrich/prmonitor
interfacer github.com/brentdrich/prmonitor
unconvert github.com/brentdrich/prmonitor