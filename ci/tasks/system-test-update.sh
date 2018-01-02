#!/bin/bash

# We can't test that concourse-up will update itself to a latest release without publishing a new release
# Instead we will test that if we publish a non-existant release, the self-update will revert back to a known release

[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; }
set -eu

deployment="system-test-$RANDOM"

cp release/concourse-up-linux-amd64 ./cup-old
cp "$BINARY_PATH" ./cup-new
chmod +x ./cup-*

echo "DEPLOY OLD VERSION"

./cup-old deploy $deployment

echo "UPDATE TO NEW VERSION"

./cup-new deploy $deployment

sleep 60

config=$(./cup-new info --region us-east-1 --json $deployment)
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

echo "DESTROYING DEPLOYMENT"
./cup-new --non-interactive destroy $deployment
