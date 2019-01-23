#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/trap.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/id.sh

[ "$VERBOSE" ] && { handleVerboseMode; }

[ -z "$SYSTEM_TEST_ID" ] && { generateSystemTestId; }
deployment="systest-$SYSTEM_TEST_ID"

set -euo pipefail

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/tags.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/credhub.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/gcreds.sh

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
    # shellcheck disable=SC2034
    region=us-east-1
    addGitHubFlagsToArgs
    addTagsFlagsToArgs
    args+=(--region "$region")
    args+=(--domain cup.engineerbetter.com)
    ./cup deploy "${args[@]}" "$deployment"
    assertTagsSet
    assertGitHubAuthConfigured

elif [ "$IAAS" = "GCP" ]
then
    # shellcheck disable=SC2034
    region=europe-west2
    gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"
    export CLOUDSDK_CORE_PROJECT=concourse-up
    addGitHubFlagsToArgs
    addTagsFlagsToArgs
    args+=(--region "$region")
    args+=(--domain cup.gcp.engineerbetter.com)
    ./cup deploy "$deployment" -iaas gcp
    assertTagsSet
    assertGitHubAuthConfigured
fi
assertPipelinesCanReadFromCredhub
sleep 60
recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy "$deployment" -iaas "$IAAS" --region "$region"
sleep 180
assertEverythingDeleted
