#!/bin/bash

set -e
gofmt -l -w .
test -z "$(gofmt -d .)"
go install github.com/brentdrich/prmonitor/cmd/prmonitor
gometalinter.v1 --deadline=30s ./...