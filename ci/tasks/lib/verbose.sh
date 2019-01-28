#!/bin/bash

function handleVerboseMode() {
  set +u
  if [ "$VERBOSE" ]; then
    set -x
    export BOSH_LOG_LEVEL=debug
    export BOSH_LOG_PATH=bosh.log
  fi
  set -u
}
