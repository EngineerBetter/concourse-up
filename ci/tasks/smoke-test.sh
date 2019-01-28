#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/set-flags.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/assert-iaas.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/id.sh

handleVerboseMode
setDeploymentName smk

cp "$BINARY_PATH" ./cup
chmod +x ./cup

./cup deploy "$deployment"
./cup --non-interactive destroy "$deployment"
