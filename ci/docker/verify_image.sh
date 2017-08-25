#!/bin/bash

set -eux

which aws
which certstrap
which go
which jq

echo "GOPATH is $GOPATH"