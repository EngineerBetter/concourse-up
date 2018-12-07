#!/bin/bash

[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; }
set -eu

deployment="systest-tags-$RANDOM"

cp "$BINARY_PATH" ./cup
chmod +x ./cup

./cup deploy $deployment
./cup --non-interactive destroy $deployment
