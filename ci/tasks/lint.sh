#!/bin/bash

set -eu

mkdir -p $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/src/bitbucket.org/engineerbetter/concourse-up

gometalinter.v1 \
  --exclude vendor \
  --exclude "_test\.go" \
  --disable=gotype \
  --disable=gas \
  --deadline=500s \
  ./...
