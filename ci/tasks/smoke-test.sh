#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/id.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/gcreds.sh

handleVerboseMode

set -e

[ -z "$SYSTEM_TEST_ID" ] && { generateSystemTestId; }
deployment="systest-$SYSTEM_TEST_ID"

set -u

cp "$BINARY_PATH" ./cup
chmod +x ./cup

# If we're testing GCP, we need credentials to be available as a file
[ "$IAAS" = "GCP" ] && { setGoogleCreds; }

./cup deploy "$deployment"
./cup --non-interactive destroy "$deployment"
