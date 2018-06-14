#!/bin/bash

[ "$VERBOSE" ] && { set -x; export BOSH_LOG_LEVEL=debug; }
set -eu

deployment="system-test-$RANDOM"

cleanup() {
  status=$?
  ./cup --non-interactive destroy --region us-east-1 $deployment
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

./cup deploy $deployment

sleep 60

eval "$(./cup info $deployment --region us-east-1 --env)"

credhub api
credhub set -n /concourse/main/backup_restore_test -t value -v kebabs

config=$(./cup info --json $deployment)
domain=$(echo "$config" | jq -r '.config.domain')
username=$(echo "$config" | jq -r '.config.concourse_username')
password=$(echo "$config" | jq -r '.config.concourse_password')

fly -t system-test login \
  --concourse-url "https://$domain" \
  --username "$username" \
  --password "$password"

cat <<EOF >pipeline.yml
---
resources:
- name: time
  type: time
  source: {interval: 5s}
jobs:
- name: get-time
  plan:
  - get: time
EOF

fly -t system-test set-pipeline -c pipeline.yml -p test
fly -t system-test unpause-pipeline -p test
sleep 30

latest_bbr_release_no=$(curl --silent "https://api.github.com/repos/cloudfoundry-incubator/bosh-backup-and-restore/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
wget https://github.com/cloudfoundry-incubator/bosh-backup-and-restore/releases/download/$latest_bbr_release_no/bbr-${latest_bbr_release_no/v}.tar
tar bbr-${latest_bbr_release_no/v}.tar

chmod +x releases/bbr
echo $BOSH_CA_CERT > bosh_ca_cert

./bbr deployment --debug --target $BOSH_ENVIRONMENT --username $BOSH_CLIENT --deployment concourse --ca-cert ./bosh_ca_cert backup --artifact-path backup_artifacts

credhub delete -n /concourse/main/backup_restore_test
fly -t system-test destroy-pipeline -p test -n
fly -t system-test destroy-team -n main --non-interactive

./bbr deployment --debug --target $BOSH_ENVIRONMENT --username $BOSH_CLIENT --deployment concourse --ca-cert ./bosh_ca_cert restore --artifact-path backup_artifacts

fly -t system-test login \
  --concourse-url "https://$domain" \
  --username "$username" \
  --password "$password"

value=$(credhub get -n /concourse/main/backup_restore_test --output-json | jq -r '.value')

[ $value == "kebabs" ]

fly -t system-test pipelines --all | grep test

### check self-backup pipeline is present and works

fly -t system-test pipelines --all | grep backup

fly -t system-test trigger-job -j concourse-up-self-backup

sleep 300

aws s3 ls $deployment-eu-west-1-config

aws s3 ls concourse-up-system-test-24496-eu-west-1-config | backup

