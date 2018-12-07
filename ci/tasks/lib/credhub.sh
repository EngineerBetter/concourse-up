#!/bin/bash

function assertPipelinesCanReadFromCredhub() {
  echo "About to test that pipelines can get values from Credhub"

  config=$(./cup info --json $deployment)
  domain=$(echo "$config" | jq -r '.config.domain')
  username=$(echo "$config" | jq -r '.config.concourse_username')
  password=$(echo "$config" | jq -r '.config.concourse_password')
  echo "$config" | jq -r '.config.concourse_ca_cert' > generated-ca-cert.pem

  eval "$(./cup info --env $deployment)"
  credhub api
  credhub set -n /concourse/main/password -t password -w c1oudc0w

  fly --target system-test login \
    --ca-cert generated-ca-cert.pem \
    --concourse-url "https://$domain" \
    --username "$username" \
    --password "$password"

  fly --target system-test sync

  fly --target system-test set-pipeline \
    --non-interactive \
    --pipeline credhub \
    --config "$(dirname "$0")/credhub.yml"

  fly --target system-test unpause-pipeline \
      --pipeline credhub

  fly --target system-test trigger-job \
    --job credhub/credhub \
    --watch

  echo "Credhub tests passed"
}
