#!/bin/bash

function handleVerboseMode() {
  set -x
  export BOSH_LOG_LEVEL=debug
  export BOSH_LOG_PATH=bosh.log
}