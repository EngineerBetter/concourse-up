#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/cleanup.sh

# if skip teardown not set, calls default cleanup
function trapDefaultCleanup() {
  set +u
  if [ -z "$SKIP_TEARDOWN" ]; then
    trap defaultCleanup EXIT
  else
    trap "echo Skipping teardown" EXIT
  fi
  set -u
}

# if skip teardown not set calls custom cleanup with region arg
function trapCustomCleanup() {
  set +u
  if [ -z "$SKIP_TEARDOWN" ]; then
    trap 'customCleanup "$region"' EXIT
  else
    trap "echo Skipping teardown" EXIT
  fi
  set -u
}
