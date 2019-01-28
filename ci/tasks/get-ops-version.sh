#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/set-flags.sh

fly -t ebci login \
  --insecure \
  --concourse-url "${CONCOURSE_URL}" \
  --username admin \
  --password "${CONCOURSE_PASSWORD}"

fly -t ebci sync

atc_bearer_token=$(bosh int --path /targets/ebci/token/value ~/.flyrc)

export ATC_BEARER_TOKEN="${atc_bearer_token}"

job=$(cat build-metadata/build-job-name)
team=$(cat build-metadata/build-team-name)

stopover https://ci.engineerbetter.com "${team}" concourse-up "${job}" "$(cat build-metadata/build-name)" > versions.yml

bosh int --path /resource_version_concourse-up-ops/ref versions.yml > ops-version/version
