#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/pipeline.sh

function assertPipelinesCanReadFromCredhub() {
  echo "About to test that pipelines can get values from Credhub"

  config=$(./cup info --region "$region" --json "$deployment" -iaas "$IAAS")
  # shellcheck disable=SC2034
  domain=$(echo "$config" | jq -r '.config.domain')
  # shellcheck disable=SC2034
  username=$(echo "$config" | jq -r '.config.concourse_username')
  # shellcheck disable=SC2034
  password=$(echo "$config" | jq -r '.config.concourse_password')
  echo "$config" | jq -r '.config.concourse_cert' > generated-ca-cert.pem

  eval "$(./cup info --env --region "$region" "$deployment" -iaas "$IAAS")"
  credhub api
  credhub set -n /concourse/main/password -t password -w c1oudc0w

  # shellcheck disable=SC2034
  manifest="$(dirname "$0")/credhub.yml"
  # shellcheck disable=SC2034
  job="credhub"
  # shellcheck disable=SC2034
  cert="generated-ca-cert.pem"

  assertPipelineIsSettableAndRunnable

  echo "Credhub tests passed"
}
