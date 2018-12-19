#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/pipeline.sh

function assertPipelinesCanReadFromCredhub() {
  echo "About to test that pipelines can get values from Credhub"

  config=$(./cup info --region us-east-1 --json "$deployment")
  domain=$(echo "$config" | jq -r '.config.domain')
  username=$(echo "$config" | jq -r '.config.concourse_username')
  password=$(echo "$config" | jq -r '.config.concourse_password')
  echo "$config" | jq -r '.config.concourse_cert' > generated-ca-cert.pem

  eval "$(./cup info --env --region us-east-1 "$deployment")"
  credhub api
  credhub set -n /concourse/main/password -t password -w c1oudc0w

  manifest="$(dirname "$0")/credhub.yml"
  job="credhub"
  cert="generated-ca-cert.pem"

  assertPipelineIsSettableAndRunnable

  echo "Credhub tests passed"
}
