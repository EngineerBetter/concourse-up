#!/bin/bash

set -eu

export GOPATH=$PWD/go
cd go/src/bitbucket.org/engineerbetter/concourse-up

go run main.go deploy system-test

hasKey=$(aws ec2 describe-key-pairs | jq -r '.KeyPairs[] | select(.KeyName | contains("engineerbetter-concourseup-system-test")) | any')

if [[ ! $hasKey == 'true' ]]; then
  echo "Couldn't find key pair stating with engineerbetter-concourseup-system-test"
  exit 1
fi

# concourse-up destroy system-test