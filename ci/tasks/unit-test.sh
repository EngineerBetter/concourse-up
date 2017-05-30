#!/bin/bash

set -eu

mkdir -p $GOPATH/src/github.com/EngineerBetter/concourse-up
mv concourse-up/* $GOPATH/src/github.com/EngineerBetter/concourse-up
cd $GOPATH/src/github.com/EngineerBetter/concourse-up

ginkgo -r
