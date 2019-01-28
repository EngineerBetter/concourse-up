#!/bin/bash

set -euo pipefail

function generateSystemTestId() {
  set +u
  if [ -z "$SYSTEM_TEST_ID" ]; then
    # ID constrained to a maximum of four characters to avoid exceeding character limit in GCP naming
    MAX_ID=9999
    SYSTEM_TEST_ID=$RANDOM
    (( SYSTEM_TEST_ID %= MAX_ID ))
  fi
  set -u
}

function setDeploymentName() {
  name=$1
  generateSystemTestId
  # shellcheck disable=SC2034
  deployment="ts-$name-$SYSTEM_TEST_ID"
}
