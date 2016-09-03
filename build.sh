#!/bin/bash

set -e
gofmt -l -w .
test -z "$(gofmt -s -d .)"

# will it build?
go install github.com/brentdrich/prmonitor/cmd/prmonitor

# will it lint?
deadcode ./..
golint ./...
errcheck ./...
interfacer ./...
unconvert ./...