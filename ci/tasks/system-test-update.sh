#!/bin/bash

# We can't test that concourse-up will update itself to a latest release without publishing a new release
# Instead we will test that if we publish a non-existant release, the self-update will revert back to a known release

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/set-flags.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/assert-iaas.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/trap.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/pipeline.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/id.sh

handleVerboseMode

setDeploymentName updt

trapDefaultCleanup

cp release/concourse-up-linux-amd64 ./cup
chmod +x ./cup

echo "DEPLOY OLD VERSION"

./cup deploy "$deployment"

# Wait for previous deployment to finish
# Otherwise terraform state can get into an invalid state
# Also wait to make sure the BOSH lock is not taken before
# starting deploy
echo "Waiting for 10 minutes to give old deploy time to settle"
sleep 600

eval "$(./cup info --env "$deployment")"
config=$(./cup info --json "$deployment")
domain=$(echo "$config" | jq -r '.config.domain')

echo "Waiting for bosh lock to become available"
wait_time=0
until [[ $(bosh locks --json | jq -r '.Tables[].Rows | length') -eq 0 ]]; do
  (( ++wait_time ))
  if [[ $wait_time -ge 10 ]]; then
    echo "Waited too long for lock" && exit 1
  fi
  printf '.'
  sleep 60
done
echo "Bosh lock available - Proceeding"

echo "UPDATE TO NEW VERSION"
# export SELF_UPDATE=true
rm -rf cup
cp "$BINARY_PATH" ./cup
chmod +x ./cup
./cup deploy "$deployment"

echo "Waiting for 30 seconds to let detached upgrade start"
sleep 30

echo "Waiting for update to complete"
wait_time=0
until curl -skIfo/dev/null "https://$domain"; do
  (( ++wait_time ))
  if [[ $wait_time -ge 10 ]]; then
    echo "Waited too long for deployment" && exit 1
  fi
  printf '.'
  sleep 30
done
echo "Update complete - Proceeding"

sleep 60

config=$(./cup info --json "$deployment")
domain=$(echo "$config" | jq -r '.config.domain')
# shellcheck disable=SC2034
username=$(echo "$config" | jq -r '.config.concourse_username')
# shellcheck disable=SC2034
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_ca_cert' > generated-ca-cert.pem

# shellcheck disable=SC2034
cert="generated-ca-cert.pem"
# shellcheck disable=SC2034
manifest="$(dirname "$0")/hello.yml"
# shellcheck disable=SC2034
job="hello"

assertPipelineIsSettableAndRunnable
