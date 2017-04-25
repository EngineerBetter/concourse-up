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

fly --target system-test login --concourse-url http://$elb_dns_name --username $username --password $password
fly --target chimichanga workers

go run main.go --non-interactive destroy $deployment

bucket="concourse-up-$deployment-config"

aws s3 rm s3://$bucket --recursive
aws s3api delete-objects --bucket $bucket --delete "$(aws s3api list-object-versions --bucket $bucket | jq -M '{Objects: [.["Versions","DeleteMarkers"][]| {Key:.Key, VersionId : .VersionId}], Quiet: false}')"
aws s3 rb s3://$bucket --force
