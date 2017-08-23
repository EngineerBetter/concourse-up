#!/bin/bash

set -eu

deployment="system-test-$RANDOM"

cp "$BINARY_PATH" ./cup-new
chmod +x ./cup-new

cp previous-release/concourse-up-linux-amd64 ./cup-old
chmod +x ./cup-old

echo "DEPLOY PREVIOUS VERSION"

./cup-old deploy $deployment

config=$(./cup-old info --region us-east-1 --json $deployment)
domain=$(echo "$config" | jq -r '.config.domain')
username=$(echo "$config" | jq -r '.config.concourse_username')
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_ca_cert' > generated-ca-cert.pem

sleep 60

fly --target system-test login \
  --ca-cert generated-ca-cert.pem \
  --concourse-url "https://$domain" \
  --username "$username" \
  --password "$password"

set -x
fly --target system-test sync
fly --target system-test workers --details
set +x

echo "DESTROY DEPLOYMENT"

./cup-old --non-interactive destroy --region us-east-1 $deployment
