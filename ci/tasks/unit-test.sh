#!/bin/bash

go get github.com/onsi/ginkgo/ginkgo github.com/onsi/gomega
mkdir -p $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
cd $GOPATH/src/bitbucket.org/engineerbetter/concourse-up
ginkgo -r
go run main.go
