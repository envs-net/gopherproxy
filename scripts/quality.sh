#!/bin/sh
set -eu

printf '%s\n' '[1/4] gofmt'
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    printf '%s\n' "$unformatted"
    printf '%s\n' 'gofmt check failed' >&2
    exit 1
fi

printf '%s\n' '[2/4] go test'
go test ./...

printf '%s\n' '[3/4] go vet'
go vet ./...

printf '%s\n' '[4/4] go build'
go build -trimpath ./cmd/gopherproxy

printf '%s\n' 'Quality checks passed.'
