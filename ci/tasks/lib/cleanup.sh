#!/bin/bash

# cleans up a deployment in the default region
function defaultCleanup() {
    status=$?
    set +e
    ./cup --non-interactive destroy "$deployment"
    set -e
    exit $status
}

# cleans up a deployment, takes region as an argument
function customCleanup() {
  status=$?
  set +e
  ./cup --non-interactive destroy "$deployment" --region "$region"
  set -e
  exit $status
}
