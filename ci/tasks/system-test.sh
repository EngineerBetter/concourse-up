#!/bin/bash

set -eu

mkdir -p $GOPATH/src/github.com/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/github.com/engineerbetter/concourse-up
cd $GOPATH/src/github.com/engineerbetter/concourse-up

deployment="system-test-$RANDOM"
domain="$deployment.concourse-up.engineerbetter.com"

go run main.go deploy $deployment --domain $domain

config=$(go run main.go info $deployment)
username=$(echo $config | jq -r '.config.concourse_username')
password=$(echo $config | jq -r '.config.concourse_password')

fly --target system-test login --insecure --concourse-url https://$domain --username $username --password $password
fly --target system-test workers

go run main.go --non-interactive destroy $deployment