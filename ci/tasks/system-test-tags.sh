#!/bin/bash

[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; }
set -eu

deployment="systest-tags-$RANDOM"

cleanup() {
  status=$?
  ./cup --non-interactive destroy $deployment
  exit $status
}
set +u
if [ -z "$SKIP_TEARDOWN" ]; then
  trap cleanup EXIT
else
  trap "echo Skipping teardown" EXIT
fi
set -u

cp "$BINARY_PATH" ./cup
chmod +x ./cup

echo "DEPLOY WITH TAGS"

./cup deploy $deployment \
  --add-tag "unique-tag=special-value" \
  --add-tag "yet-another-tag=some-value"

tagged_instances=$(aws ec2 --region eu-west-1 \
  describe-instances \
  --filters="Name=tag:unique-tag,Values=special-value,Name=tag:yet-another-tag,Values=some-value,Name=tag:concourse-up-project,Values=$deployment" \
  | jq -r '.Reservations | length')

if [[ $tagged_instances -ne 3 ]]; then
  echo "Expected 3 tagged instances, got $tagged_instances"
  exit 1
fi

echo "Test Successful"
