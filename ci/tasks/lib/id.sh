#!/bin/bash

function generateSystemTestId() {
  # ID constrained to a maximum of four characters to avoid exceeding character limit in GCP naming
  MAX_ID=9999
  SYSTEM_TEST_ID=$RANDOM
  (( SYSTEM_TEST_ID %= MAX_ID ))
}