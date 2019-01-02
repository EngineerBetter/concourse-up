#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/trap.sh

[ "$VERBOSE" ] && { handleVerboseMode; }

set -euo pipefail

deployment="systest-github-$RANDOM"

set +u
# shellcheck disable=SC2034
region=us-east-1
trapCustomCleanup us-east-1
set -u

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/tags.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/credhub.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/destroy.sh


cp "$BINARY_PATH" ./cup
chmod +x ./cup

addGitHubFlagsToArgs
addTagsFlagsToArgs
./cup deploy "${args[@]}" $deployment
assertTagsSet
assertGitHubAuthConfigured
assertPipelinesCanReadFromCredhub
sleep 60
recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy --region us-east-1 "$deployment"
sleep 180
assertEverythingDeleted
