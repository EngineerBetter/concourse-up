#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/trap.sh

[ "$VERBOSE" ] && { handleVerboseMode; }

set -euo pipefail

deployment="systest-github-$RANDOM"

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/tags.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/credhub.sh

# shellcheck disable=SC1091
[ "$IAAS" = "AWS" ] && { source concourse-up/ci/tasks/lib/destroy.sh; }

# shellcheck disable=SC1091
[ "$IAAS" = "GCP" ] && { source concourse-up/ci/tasks/lib/gcp-destroy.sh; }

# If we're testing GCP, we need credentials to be available as a file
[ "$IAAS" = "GCP" ] && { setGoogleCreds; }


cp "$BINARY_PATH" ./cup
chmod +x ./cup

# Temporary fix whilst we haven't implemented the rest for GCP
if [ "$IAAS" = "AWS" ]
then
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
elif [ "$IAAS" = "GCP" ]
then
    ./cup deploy $deployment -iaas gcp
    sleep 60
    recordDeployedState
    echo "non-interactive destroy"
    ./cup --non-interactive destroy "$deployment" -iaas gcp
fi
sleep 180
assertEverythingDeleted