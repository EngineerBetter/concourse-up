#!/bin/bash

set -eu

build_dir=$PWD/build
mkdir -p build_dir

mkdir -p $GOPATH/src/github.com/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/github.com/engineerbetter/concourse-up
cd $GOPATH/src/github.com/engineerbetter/concourse-up

GOOS=linux GOARCH=amd64 go build -o $build_dir/concourse-up-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o $build_dir/concourse-up-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o $build_dir/concourse-up-windows-amd64.exe
