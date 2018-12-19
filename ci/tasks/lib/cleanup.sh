#!/bin/bash

# cleans up a deployment in the default region
function defaultCleanup() {
    status=$?
    ./cup --non-interactive destroy "$deployment"
    exit $status
}

# cleans up a deployment, takes region as an argument
function customCleanup() {
  status=$?
  ./cup --non-interactive destroy "$deployment" --region "$region"
  exit $status
}