#!/bin/bash

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/verbose.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/trap.sh

# shellcheck disable=SC1091
source concourse-up/ci/tasks/lib/pipeline.sh

[ "$VERBOSE" ] && { handleVerboseMode; }

set -eu

deployment="systest-region-$RANDOM"

set +u

trapCustomCleanup eu-west-3

set -u

cp "$BINARY_PATH" ./cup
chmod +x ./cup

echo "DEPLOY WITH AUTOGENERATED CERT, NO DOMAIN, CUSTOM REGION, DEFAULT WORKERS"
./cup deploy --worker-type m5 --region eu-west-3 $deployment

sleep 60

config=$(./cup info --region eu-west-3 --json $deployment)
# shellcheck disable=SC2034
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