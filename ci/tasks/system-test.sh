#!/bin/bash

set -eux

mkdir -p $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/src/bitbucket.org/engineerbetter/concourse-up

deployment="system-test-$RANDOM"

go run main.go deploy $deployment

config=$(go run main.go info $deployment)
director_ip=$(echo $config | jq -r '.terraform.director_public_ip.value')
username=$(echo $config | jq -r '.config.director_username')
password=$(echo $config | jq -r '.config.director_password')
cacert="$(echo $config | jq -r '.config.director_ca_cert')"

echo "$cacert" > cacert.pem

bosh-cli \
  --ca-cert=cacert.pem \
  --environment $director_ip \
  --client $username \
  --client-secret $password \
  deployments

go run main.go --non-interactive destroy $deployment

bucket="concourse-up-$deployment"

aws s3 rm s3://$bucket --recursive
aws s3api delete-objects --bucket $bucket --delete "$(aws s3api list-object-versions --bucket $bucket | jq -M '{Objects: [.["Versions","DeleteMarkers"][]| {Key:.Key, VersionId : .VersionId}], Quiet: false}')"
aws s3 rb s3://$bucket --force
