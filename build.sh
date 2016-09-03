#!/bin/bash

set -e
gofmt -l -w .
test -z "$(gofmt -d .)"
go build
go clean
gometalinter.v1 --deadline=30s ./...