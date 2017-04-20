#!/bin/bash

set -eu

mkdir -p $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/src/bitbucket.org/engineerbetter/concourse-up

gometalinter.v1 -e vendor -e "_test\.go" --deadline=120s ./...
