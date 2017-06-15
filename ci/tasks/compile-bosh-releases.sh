#!/bin/bash

set -eu

echo "$BOSH_CA_CERT" > bosh_ca_cert.pem

bosh_flags="--non-interactive --environment $BOSH_TARGET --client $BOSH_USERNAME --client-secret $BOSH_PASSWORD --ca-cert bosh_ca_cert.pem"

concourse_stemcell_version=$(cat concourse-stemcell/version)
concourse_stemcell_url=$(cat concourse-stemcell/url)
concourse_stemcell_sha1=$(cat concourse-stemcell/sha1)

director_stemcell_version=$(cat director-stemcell/version)
director_stemcell_url=$(cat director-stemcell/url)
director_stemcell_sha1=$(cat director-stemcell/sha1)

director_bosh_release_version=$(cat director-bosh-release/version)
director_bosh_cpi_release_version=$(cat director-bosh-cpi-release/version)
concourse_release_version=$(ls concourse-bosh-release/concourse-*.tgz | awk -F"-" '{ print $4 }' | awk -F".tgz" '{ print $1 }')
garden_release_version=$(ls concourse-bosh-release/garden-runc-*.tgz | awk -F"-" '{ print $5 }' | awk -F".tgz" '{ print $1 }')

bosh-cli $bosh_flags upload-stemcell "concourse-stemcell/stemcell.tgz"
bosh-cli $bosh_flags upload-release "concourse-bosh-release/garden-runc-$garden_release_version.tgz"
bosh-cli $bosh_flags upload-release "concourse-bosh-release/concourse-$concourse_release_version.tgz"
bosh-cli $bosh_flags upload-release "director-bosh-release/release.tgz"
bosh-cli $bosh_flags upload-release "director-bosh-cpi-release/release.tgz"

echo "---
name: concourse-empty

releases:
- name: concourse
  version: \"$concourse_release_version\"
- name: garden-runc
  version: \"$garden_release_version\"
- name: bosh
  version: \"$director_bosh_release_version\"
- name: bosh-aws-cpi
  version: \"$director_bosh_cpi_release_version\"

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: \"$concourse_stemcell_version\"

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
  export-release concourse/$concourse_release_version ubuntu-trusty/$concourse_stemcell_version

bosh-cli $bosh_flags \
  --deployment concourse-empty \
  export-release garden-runc/$garden_release_version ubuntu-trusty/$concourse_stemcell_version

bosh-cli $bosh_flags \
  --deployment concourse-empty \
  export-release bosh/$director_bosh_release_version ubuntu-trusty/$concourse_stemcell_version

bosh-cli $bosh_flags \
  --deployment concourse-empty \
  export-release bosh-aws-cpi/$director_bosh_cpi_release_version ubuntu-trusty/$concourse_stemcell_version

compiled_concourse_release=$(ls concourse-$concourse_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_garden_release=$(ls garden-runc-$garden_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_director_bosh_release=$(ls bosh-$director_bosh_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_director_bosh_cpi_release=$(ls bosh-aws-cpi-$director_bosh_cpi_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)

aws s3 cp --acl public-read "$compiled_concourse_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_concourse_release"
aws s3 cp --acl public-read "$compiled_garden_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_garden_release"
aws s3 cp --acl public-read "$compiled_director_bosh_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_release"
aws s3 cp --acl public-read "$compiled_director_bosh_cpi_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_cpi_release"

director_bosh_release_sha1=$(sha1sum "$compiled_director_bosh_release" | awk '{ print $1 }')
director_bosh_cpi_release_sha1=$(sha1sum "$compiled_director_bosh_cpi_release" | awk '{ print $1 }')
concourse_release_sha1=$(sha1sum "$compiled_concourse_release" | awk '{ print $1 }')
garden_release_sha1=$(sha1sum "$compiled_garden_release" | awk '{ print $1 }')

echo "{
  \"concourse_stemcell_url\": \"$concourse_stemcell_url\",
  \"concourse_stemcell_sha1\": \"$concourse_stemcell_sha1\",
  \"concourse_stemcell_version\": \"$concourse_stemcell_version\",

  \"director_stemcell_url\": \"$director_stemcell_url\",
  \"director_stemcell_sha1\": \"$director_stemcell_sha1\",
  \"director_stemcell_version\": \"$director_stemcell_version\",

  \"director_bosh_release_url\": \"https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_release\",
  \"director_bosh_release_sha1\": \"$director_bosh_release_sha1\",
  \"director_bosh_release_version\": \"$director_bosh_release_version\",

  \"director_bosh_cpi_release_url\": \"https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_cpi_release\",
  \"director_bosh_cpi_release_sha1\": \"$director_bosh_cpi_release_sha1\",
  \"director_bosh_cpi_release_version\": \"$director_bosh_cpi_release_version\",

  \"concourse_release_url\": \"https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_concourse_release\",
  \"concourse_release_sha1\": \"$concourse_release_sha1\",
  \"concourse_release_version\": \"$concourse_release_version\",

  \"garden_release_url\": \"https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_garden_release\",
  \"garden_release_sha1\": \"$garden_release_sha1\",
  \"garden_release_version\": \"$garden_release_version\"
}" > compilation-vars/compilation-vars.json