#!/bin/bash

set -eu

echo "$BOSH_CA_CERT" > bosh_ca_cert.pem

bosh_flags="--non-interactive --environment $BOSH_TARGET --client $BOSH_USERNAME --client-secret $BOSH_PASSWORD --ca-cert bosh_ca_cert.pem"

stemcell_version=$(cat concourse-stemcell/version)
concourse_release_version=$(ls concourse-bosh-release/concourse-*.tgz | awk -F"-" '{ print $4 }' | awk -F".tgz" '{ print $1 }')
garden_release_version=$(ls concourse-bosh-release/garden-runc-*.tgz | awk -F"-" '{ print $5 }' | awk -F".tgz" '{ print $1 }')

bosh-cli $bosh_flags upload-stemcell concourse-stemcell/stemcell.tgz
bosh-cli $bosh_flags upload-release concourse-bosh-release/garden-runc-$garden_release_version.tgz
bosh-cli $bosh_flags upload-release concourse-bosh-release/concourse-$concourse_release_version.tgz

echo "---
name: concourse-empty

releases:
- name: concourse
  version: $concourse_release_version
- name: garden-runc
  version: $garden_release_version

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: $stemcell_version

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000" > concourse-empty.yml

bosh-cli $bosh_flags \
  --deployment concourse-empty \
  deploy \
  concourse-empty.yml

bosh-cli $bosh_flags \
  --deployment concourse-empty \
  export-release concourse/$concourse_release_version ubuntu-trusty/$stemcell_version

bosh-cli $bosh_flags \
  --deployment concourse-empty \
  export-release garden-runc/$garden_release_version ubuntu-trusty/$stemcell_version

compiled_concourse_release=$(ls concourse-$concourse_release_version-ubuntu-trusty-$stemcell_version-*.tgz)
compiled_garden_release=$(ls garden-runc-$garden_release_version-ubuntu-trusty-$stemcell_version-*.tgz)

aws s3 cp --acl public-read $compiled_concourse_release s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_concourse_release
aws s3 cp --acl public-read $compiled_garden_release s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_garden_release

stemcell_url=$(cat concourse-stemcell/url)

echo "{
  \"stemcell_url\": \"$stemcell_url\",
  \"concourse_release_url\": \"https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_concourse_release\",
  \"garden_release_url\": \"https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_garden_release\"
}" > compilation-vars/compilation-vars.json