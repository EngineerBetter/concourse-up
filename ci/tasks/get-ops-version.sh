#!/bin/bash

set -eu

fly -t ebci login \
  --insecure \
  --concourse-url $CONCOURSE_URL \
  --username admin \
  --password $CONCOURSE_PASSWORD

fly -t ebci sync

export ATC_BEARER_TOKEN=$(bosh int --path /targets/ebci/token/value ~/.flyrc)

job=$(cat build-metadata/build-job-name)
team=$(cat build-metadata/build-team-name)

stopover https://ci.engineerbetter.com $team concourse-up $job $(cat build-metadata/build-name) > versions.yml

bosh int --path /resource_version_concourse-up-ops/ref versions.yml > ops-version/version
