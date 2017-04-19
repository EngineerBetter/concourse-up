#!/bin/bash

set -eu

export GOPATH=$PWD/go
cd go/src/bitbucket.org/engineerbetter/concourse-up

ginkgo -r
