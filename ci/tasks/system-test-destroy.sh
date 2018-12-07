#!/bin/bash

set -e
[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; export BOSH_LOG_PATH=bosh.log; }
if [ -z "$SYSTEM_TEST_ID" ]; then
    SYSTEM_TEST_ID=$RANDOM
fi
deployment="systest-cleanup-$SYSTEM_TEST_ID"
set -u

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/destroy.sh

cp "$BINARY_PATH" ./cup
chmod +x ./cup

./cup deploy --region us-east-1 $deployment
sleep 60
recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy --region us-east-1 $deployment
sleep 180
assertEverythingDeleted
