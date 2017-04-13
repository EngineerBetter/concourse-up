#!/bin/bash

set -eu

mkdir -p $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/src/bitbucket.org/engineerbetter/concourse-up

go get github.com/onsi/ginkgo/ginkgo github.com/onsi/gomega

ginkgo -r
go run main.go
