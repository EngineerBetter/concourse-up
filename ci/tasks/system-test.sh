#!/bin/bash

set -eu

mkdir -p $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/src/bitbucket.org/engineerbetter/concourse-up

deployment="system-test-$RANDOM"

go run main.go deploy $deployment

config=$(go run main.go info $deployment)
elb_dns_name=$(echo $config | jq -r '.terraform.elb_dns_name.value')
username=$(echo $config | jq -r '.config.concourse_username')
password=$(echo $config | jq -r '.config.concourse_password')

fly --target system-test login --insecure --concourse-url https://$elb_dns_name --username $username --password $password
fly --target system-test workers

go run main.go --non-interactive destroy $deployment