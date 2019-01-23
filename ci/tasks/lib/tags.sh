#!/bin/bash

# Add flags to an array that should have been initialised previously
function addTagsFlagsToArgs() {
  args+=(--add-tag "unique-tag=special-value")
  args+=(--add-tag "yet-another-tag=some-value")
}

function assertTagsSet() {
  echo "About to test that VMs are tagged"

  tagged_instances=0

  if [ "$IAAS" = "GCP" ]
  then
    tagged_instances=$(gcloud compute instances list --filter="labels.unique-tag:special-value AND labels.yet-another-tag:some-value" | awk 'NR>1 {print}'| wc -l)
  elif [ "$IAAS" = "AWS" ]
  then
    tagged_instances=$(aws ec2 --region us-east-1 \
        describe-instances \
        --filters="Name=tag:unique-tag,Values=special-value,Name=tag:yet-another-tag,Values=some-value,Name=tag:concourse-up-project,Values=$deployment" \
        | jq -r '.Reservations | length')
  fi

  if [[ $tagged_instances -ne 3 ]]; then
    echo "Expected 3 tagged instances, got $tagged_instances"
    exit 1
  fi

  echo "Tags test passed"
}
