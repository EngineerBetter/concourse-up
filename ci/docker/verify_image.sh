#!/bin/bash

set -eux

command -v aws
command -v certstrap
command -v go
command -v jq
command -v bosh-cli

echo "GOPATH is $GOPATH"