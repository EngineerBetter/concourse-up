#!/bin/bash

# Add flags to an array that should have been initialised previously
function addTagsFlagsToArgs() {
  args+=(--add-tag "unique-tag=special-value")
  args+=(--add-tag "yet-another-tag=some-value")
}

function assertTagsSet() {
  echo "About to test that VMs are tagged"

  tagged_instances=$(aws ec2 --region us-east-1 \
    describe-instances \
    --filters="Name=tag:unique-tag,Values=special-value,Name=tag:yet-another-tag,Values=some-value,Name=tag:concourse-up-project,Values=$deployment" \
    | jq -r '.Reservations | length')

  if [[ $tagged_instances -ne 3 ]]; then
    echo "Expected 3 tagged instances, got $tagged_instances"
    exit 1
  fi

  echo "Tags test passed"
}
