#!/bin/bash

: "${IAAS:=AWS}"

set -e

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/id.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/trap.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/gcreds.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/pipeline.sh

[ "$VERBOSE" ] && { handleVerboseMode; }

[ -z "$SYSTEM_TEST_ID" ] && { generateSystemTestId; }

deployment="systest-$SYSTEM_TEST_ID"

set -u

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/check-db.sh

# If we're testing GCP, we need credentials to be available as a file
[ "$IAAS" = "GCP" ] && { setGoogleCreds; }

set +u
trapDefaultCleanup
set -u

cp "$BINARY_PATH" ./cup
chmod +x ./cup


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
   --domain "$custom_domain"

certstrap sign "$custom_domain" --CA "$deployment"

echo "DEPLOY WITH A USER PROVIDED CERT, CUSTOM DOMAIN, DEFAULT WORKERS, DEFAULT DATABASE SIZE AND DEFAULT WEB NODE SIZE"

./cup deploy "$deployment" \
  --domain "$custom_domain" \
  --tls-cert "$(cat out/"$custom_domain".crt)" \
  --tls-key "$(cat out/"$custom_domain".key)"

if [ "$IAAS" = "GCP" ]; then
  echo "Testing GCP, exiting early"
  exit 1
fi
sleep 60


# Check we can log into the BOSH director and SSH into a VM
eval "$(./cup info --env "$deployment")"
bosh vms
bosh ssh worker true

config=$(./cup info --json "$deployment")
# shellcheck disable=SC2034
username=$(echo "$config" | jq -r '.config.concourse_username')
# shellcheck disable=SC2034
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_cert' > generated-ca-cert.pem


# Check RDS instance class is db.t2.small
assertDbCorrect

# shellcheck disable=SC2034
cert="generated-ca-cert.pem"
# shellcheck disable=SC2034
manifest="$(dirname "$0")/hello.yml"
# shellcheck disable=SC2034
job="hello"
# shellcheck disable=SC2034
domain=$custom_domain

assertPipelineIsSettableAndRunnable


echo "DEPLOY 2 LARGE WORKERS, FIREWALLED TO MY IP"


./cup deploy "$deployment" \
  --allow-ips "$(dig +short myip.opendns.com @resolver1.opendns.com)" \
  --workers 2 \
  --worker-size large

sleep 60

# Check RDS instance class is still db.t2.small
assertDbCorrect

config=$(./cup info --json "$deployment")
# shellcheck disable=SC2034
username=$(echo "$config" | jq -r '.config.concourse_username')
# shellcheck disable=SC2034
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_ca_cert' > generated-ca-cert.pem
# shellcheck disable=SC2034
cert="generated-ca-cert.pem"

assertPipelineIsRunnable