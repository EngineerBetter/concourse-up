#!/bin/bash

set -eu

dirname=$(dirname $0)
. $dirname/../script_setup.sh

go get github.com/onsi/ginkgo/ginkgo github.com/onsi/gomega

ginkgo -r
go run main.go
