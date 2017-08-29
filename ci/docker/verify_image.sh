#!/bin/bash

set -eux

which aws
which certstrap
which go
which jq
which bosh-cli

echo "GOPATH is $GOPATH"