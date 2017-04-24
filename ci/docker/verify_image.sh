#!/bin/bash

set -eux

which aws
which terraform
which go
which jq
which bosh-cli

echo "GOPATH is $GOPATH"