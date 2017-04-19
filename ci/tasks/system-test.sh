#!/bin/bash

set -eu

export GOPATH=$PWD/go
cd go/src/bitbucket.org/engineerbetter/concourse-up

deployment="engineerbetter-concourseup-system-test-$RANDOM"

go run main.go deploy $deployment

hasKey=$(aws ec2 describe-key-pairs | jq -r ".KeyPairs[] | select(.KeyName | contains(\"$deployment\")) | any")

if [[ ! $hasKey == 'true' ]]; then
  echo "Couldn't find key pair stating with $deployment"
  exit 1
fi

concourse-up --non-interactive destroy $deployment

aws s3api delete-objects --bucket $deployment --delete "$(aws s3api list-object-versions --bucket $deployment | jq -M '{Objects: [.["Versions","DeleteMarkers"][]| {Key:.Key, VersionId : .VersionId}], Quiet: false}')"

aws s3 rb s3://$deployment --force
