#!/bin/bash

# We can't test that concourse-up will update itself to a latest release without publishing a new release
# Instead we will test that if we publish a non-existant release, the self-update will revert back to a known release

[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; }
set -eu

deployment="systest-update-$RANDOM"

cleanup() {
  status=$?
  ./cup-new --non-interactive destroy $deployment
  exit $status
}
set +u
if [ -z "$SKIP_TEARDOWN" ]; then
  trap cleanup EXIT
else
  trap "echo Skipping teardown" EXIT
fi
set -u

cp release/concourse-up-linux-amd64 ./cup-old
cp "$BINARY_PATH" ./cup-new
chmod +x ./cup-*

echo "DEPLOY OLD VERSION"

./cup-old deploy $deployment

# Wait for previous deployment to finish
# Otherwise terraform state can get into an invalid state
# Also wait to make sure the BOSH lock is not taken before
# starting deploy
echo "Waiting for 10 minutes to give old deploy time to settle"
sleep 600

eval "$(./cup-old info --env $deployment)"
config=$(./cup-old info --json $deployment)
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

./cup-new deploy $deployment

echo "Waiting for 30 seconds to let detached upgrade start"
sleep 30

echo "Waiting for update to complete"
wait_time=0
# shellcheck disable=SC2091
until $(curl -skIfo/dev/null "https://$domain"); do
  (( ++wait_time ))
  if [[ $wait_time -ge 10 ]]; then
    echo "Waited too long for deployment" && exit 1
  fi
  printf '.'
  sleep 30
done
echo "Update complete - Proceeding"

sleep 60

config=$(./cup-new info --json $deployment)
domain=$(echo "$config" | jq -r '.config.domain')
username=$(echo "$config" | jq -r '.config.concourse_username')
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_ca_cert' > generated-ca-cert.pem

fly --target system-test login \
  --ca-cert generated-ca-cert.pem \
  --concourse-url "https://$domain" \
  --username "$username" \
  --password "$password"

curl -k "https://$domain:3000"

fly --target system-test sync

fly --target system-test set-pipeline \
  --non-interactive \
  --pipeline hello \
  --config "$(dirname "$0")/hello.yml"

fly --target system-test unpause-pipeline \
    --pipeline hello

fly --target system-test trigger-job \
  --job hello/hello \
  --watch
