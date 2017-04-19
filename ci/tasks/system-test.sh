#!/bin/bash

set -eux

export GOPATH=$PWD/go
cd go/src/bitbucket.org/engineerbetter/concourse-up

deployment="system-test-$RANDOM"

go run main.go deploy $deployment

hasKey=$(aws ec2 describe-key-pairs | jq -r ".KeyPairs[] | select(.KeyName | contains(\"$deployment\")) | any")

if [[ ! $hasKey == 'true' ]]; then
  echo "Couldn't find key pair starting with $deployment"
  exit 1
fi

go run main.go --non-interactive destroy $deployment

bucket="engineerbetter-concourseup-$deployment"

aws s3 rm s3://$bucket --recursive
aws s3api delete-objects --bucket $bucket --delete "$(aws s3api list-object-versions --bucket $bucket | jq -M '{Objects: [.["Versions","DeleteMarkers"][]| {Key:.Key, VersionId : .VersionId}], Quiet: false}')"
aws s3 rb s3://$bucket --force
