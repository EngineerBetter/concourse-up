#!/bin/bash

set -eu

export GOPATH=$PWD/go
cd go/src/bitbucket.org/engineerbetter/concourse-up

go run main.go deploy system-test

bucket=engineerbetter-concourseup-system-test

hasKey=$(aws ec2 describe-key-pairs | jq -r '.KeyPairs[] | select(.KeyName | contains("$bucket")) | any')

if [[ ! $hasKey == 'true' ]]; then
  echo "Couldn't find key pair stating with $bucket"
  exit 1
fi

concourse-up --non-interactive destroy system-test

aws s3api delete-objects --bucket $bucket --delete "$(aws s3api list-object-versions --bucket $bucket | jq -M '{Objects: [.["Versions","DeleteMarkers"][]| {Key:.Key, VersionId : .VersionId}], Quiet: false}')"

aws s3 rb s3://$bucket --force