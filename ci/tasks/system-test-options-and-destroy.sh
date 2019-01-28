#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/set-flags.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/assert-iaas.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/trap.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/id.sh

handleVerboseMode
setDeploymentName opt

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/tags.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/credhub.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/letsencrypt.sh

# shellcheck disable=SC1091
[ "$IAAS" = "AWS" ] && { source concourse-up/ci/tasks/lib/destroy.sh; }

# shellcheck disable=SC1091
[ "$IAAS" = "GCP" ] && { source concourse-up/ci/tasks/lib/gcp-destroy.sh; }

cp "$BINARY_PATH" ./cup
chmod +x ./cup

if [ "$IAAS" = "AWS" ]
then
    # shellcheck disable=SC2034
    region=us-east-1
    args+=(--domain cup.engineerbetter.com)
elif [ "$IAAS" = "GCP" ]
then
    # shellcheck disable=SC2034
    region=europe-west2
    gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"
    export CLOUDSDK_CORE_PROJECT=concourse-up
    args+=(--domain cup.gcp.engineerbetter.com)
fi

addGitHubFlagsToArgs
addTagsFlagsToArgs
args+=(--region "$region")
./cup deploy "${args[@]}" --iaas "$IAAS" "$deployment"
assertTagsSet
assertGitHubAuthConfigured

assertPipelinesCanReadFromCredhub
sleep 60
recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy "$deployment" -iaas "$IAAS" --region "$region"
sleep 180
assertEverythingDeleted
