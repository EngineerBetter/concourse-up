#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/test-setup.sh

handleVerboseMode
setDeploymentName opt

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/tags.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/letsencrypt.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/assert-iaas.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/check-cidr-ranges.sh

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

    args+=(--vpc-network-range 192.168.0.0/24)
    args+=(--rds-subnet-range1 192.168.0.64/28)
    args+=(--rds-subnet-range2 192.168.0.80/28)
elif [ "$IAAS" = "GCP" ]
then
    # shellcheck disable=SC2034
    region=europe-west2
    args+=(--domain cup.gcp.engineerbetter.com)
fi

args+=(--public-subnet-range 192.168.0.0/27)
args+=(--private-subnet-range 192.168.0.32/27)

trapCustomCleanup

addGitHubFlagsToArgs
addTagsFlagsToArgs
args+=(--region "$region")
./cup deploy "${args[@]}" --iaas "$IAAS" "$deployment"
assertTagsSet
assertGitHubAuthConfigured

if [ "$IAAS" = "AWS" ]
then
    assertNetworkCidrsCorrect 192.168.0.0/27 192.168.0.32/27 192.168.0.0/24 192.168.0.64/28 192.168.0.80/28
elif [ "$IAAS" = "GCP" ]
then
    assertNetworkCidrsCorrect 192.168.0.0/27 192.168.0.32/27
fi

assertPipelinesCanReadFromCredhub
sleep 60
recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy "$deployment" -iaas "$IAAS" --region "$region"
sleep 180
assertEverythingDeleted
