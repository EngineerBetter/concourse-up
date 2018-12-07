#!/bin/bash

[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; }
set -euo pipefail

deployment="systest-github-$RANDOM"

cleanup() {
  status=$?
  ./cup --non-interactive destroy --region us-east-1 $deployment
  exit $status
}
set +u
if [ -z "$SKIP_TEARDOWN" ]; then
  trap cleanup EXIT
else
  trap "echo Skipping teardown" EXIT
fi
set -u

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/tags.sh
# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/credhub.sh

cp "$BINARY_PATH" ./cup
chmod +x ./cup

addGitHubFlagsToArgs
addTagsFlagsToArgs
./cup deploy "${args[@]}" $deployment
assertTagsSet
assertGitHubAuthConfigured
assertPipelinesCanReadFromCredhub
