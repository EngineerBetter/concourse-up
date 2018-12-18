#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/handleVerboseMode.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/generateSystemTestId.sh


[ "$VERBOSE" ] && { handleVerboseMode; }

set -e


[ -z "$SYSTEM_TEST_ID" ] && { generateSystemTestId; }
deployment="systest-$SYSTEM_TEST_ID"

set -u

cp "$BINARY_PATH" ./cup
chmod +x ./cup

# If we're testing GCP, we need credentials to be available as a file
if [ "$IAAS" = "GCP" ]; then
  echo "${GOOGLE_APPLICATION_CREDENTIALS_CONTENTS}" > googlecreds.json
  export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
fi

./cup deploy "$deployment"
./cup --non-interactive destroy "$deployment"
