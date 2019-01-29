#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/test-setup.sh

handleVerboseMode
setDeploymentName smk
trapDefaultCleanup

cp "$BINARY_PATH" ./cup
chmod +x ./cup

./cup deploy "$deployment"
./cup --non-interactive destroy "$deployment"
