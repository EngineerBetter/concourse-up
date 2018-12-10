#!/bin/bash

[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; }
set -e

if [ -z "$SYSTEM_TEST_ID" ]; then
  # ID constrained to a maximum of four characters to avoid exceeding character limit in GCP naming
  MAX_ID=9999
  SYSTEM_TEST_ID=$RANDOM
  (( SYSTEM_TEST_ID %= MAX_ID ))
fi
deployment="systest-$SYSTEM_TEST_ID"

set -u

cp "$BINARY_PATH" ./cup
chmod +x ./cup

# If we're testing GCP, we need credentials to be available as a file
if [ "$IAAS" = "GCP" ]; then
  echo "${GOOGLE_APPLICATION_CREDENTIALS_CONTENTS}" > googlecreds.json
  export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
fi

./cup deploy $deployment
./cup --non-interactive destroy $deployment
