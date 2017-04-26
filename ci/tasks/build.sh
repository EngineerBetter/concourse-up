#!/bin/bash

set -eu

mkdir -p $GOPATH/src/github.com/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/github.com/engineerbetter/concourse-up
cd $GOPATH/src/github.com/engineerbetter/concourse-up

GOOS=linux GOARCH=amd64 go build -o build/concourse-up-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o build/concourse-up-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o build/concourse-up-windows-amd64.exe
