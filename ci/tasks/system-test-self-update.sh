#!/bin/bash

# We can't test that concourse-up will update itself to a latest release without publishing a new release
# Instead we will test that if we publish a non-existant release, the self-update will revert back to a known release

set -eu

deployment="system-test-$RANDOM"

cp "$BINARY_PATH" ./cup-new
chmod +x ./cup-new

echo "DEPLOY NEW VERSION"

custom_domain="$deployment-user.concourse-up.engineerbetter.com"

certstrap init \
  --common-name "$deployment" \
  --passphrase "" \
  --organization "" \
  --organizational-unit "" \
  --country "" \
  --province "" \
  --locality ""

certstrap request-cert \
   --passphrase "" \
   --domain $custom_domain

certstrap sign "$custom_domain" --CA "$deployment"

./cup-new deploy $deployment \
  --region eu-west-2 \
  --domain $custom_domain \
  --tls-cert "$(cat out/$custom_domain.crt)" \
  --tls-key "$(cat out/$custom_domain.key)"

sleep 60

config=$(./cup-new info --region eu-west-2 --json $deployment)
domain=$(echo "$config" | jq -r '.config.domain')
username=$(echo "$config" | jq -r '.config.concourse_username')
password=$(echo "$config" | jq -r '.config.concourse_password')

fly --target system-test login \
  --ca-cert out/$deployment.crt \
  --concourse-url "https://$domain" \
  --username "$username" \
  --password "$password"

curl -k "https://$domain:3000"

set -x
fly --target system-test sync
fly --target system-test workers --details
set +x

echo "TRIGGERING SELF-UPDATE"
fly --target system-test trigger-job -j concourse-up-self-update/self-update

while true;
do
  builds="$(fly --target system-test builds)"
  echo "$builds"

  if [[ $builds != *"concourse-up-self-update/self-update"* ]] ; then
    echo "could't find self-update job in builds:\n$builds"
    exit 1
  fi

  if [[ $builds == *"succeeded"* ]] ; then
    break
  fi

  if [[ $builds == *"failed"* ]] ; then
    echo "build failed:\n$builds"
    exit 1
  fi

  sleep 1;
done

echo "DESTROYING DEPLOYMENT"
./cup-new --non-interactive destroy --region eu-west-2 $deployment
