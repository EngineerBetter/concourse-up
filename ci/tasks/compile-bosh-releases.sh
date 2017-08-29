#!/bin/bash

set -eu


# delete any concourse versions that have pre-compiled packages
rm concourse-bosh-release/concourse-*-*.tgz

echo "$BOSH_CA_CERT" > bosh_ca_cert.pem

bosh="bosh-cli --non-interactive --environment $BOSH_TARGET --client $BOSH_USERNAME --client-secret $BOSH_PASSWORD --ca-cert bosh_ca_cert.pem"

concourse_stemcell_version=$(cat concourse-stemcell/version)
concourse_stemcell_url=$(cat concourse-stemcell/url)
concourse_stemcell_sha1=$(cat concourse-stemcell/sha1)

director_stemcell_version=$(cat director-stemcell/version)
director_stemcell_url=$(cat director-stemcell/url)
director_stemcell_sha1=$(cat director-stemcell/sha1)

director_bosh_cpi_release_version=$(cat director-bosh-cpi-release/version)
director_bosh_cpi_release_url=$(cat director-bosh-cpi-release/url)
director_bosh_cpi_release_sha1=$(cat director-bosh-cpi-release/sha1)

riemann_release_version=$(cat riemann-release/version)
riemann_release_url=$(cat riemann-release/url)
riemann_release_sha1=$(cat riemann-release/sha1)

grafana_release_version=$(cat grafana-release/version)
grafana_release_url=$(cat grafana-release/url)
grafana_release_sha1=$(cat grafana-release/sha1)

influxdb_release_version=$(cat influxdb-release/version)
influxdb_release_url=$(cat influxdb-release/url)
influxdb_release_sha1=$(cat influxdb-release/sha1)

director_bosh_release_version=$(cat director-bosh-release/version)
concourse_release_version=$(ls concourse-bosh-release/concourse-*.tgz | awk -F"-" '{ print $4 }' | awk -F".tgz" '{ print $1 }')
garden_release_version=$(ls concourse-bosh-release/garden-runc-*.tgz | awk -F"-" '{ print $5 }' | awk -F".tgz" '{ print $1 }')

$bosh upload-stemcell "concourse-stemcell/stemcell.tgz"
$bosh upload-release "concourse-bosh-release/garden-runc-$garden_release_version.tgz"
$bosh upload-release "concourse-bosh-release/concourse-$concourse_release_version.tgz"
$bosh upload-release "director-bosh-release/release.tgz"
$bosh upload-release "director-bosh-cpi-release/release.tgz"
$bosh upload-release "riemann-release/release.tgz"
$bosh upload-release "grafana-release/release.tgz"
$bosh upload-release "influxdb-release/release.tgz"

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
- name: riemann
  version: \"$riemann_release_version\"
- name: grafana
  version: \"$grafana_release_version\"
- name: influxdb
  version: \"$influxdb_release_version\"

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: \"$concourse_stemcell_version\"

jobs: []

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000" > concourse-empty.yml

$bosh \
  --deployment concourse-empty \
  deploy \
  concourse-empty.yml

$bosh \
  --deployment concourse-empty \
  export-release "concourse/$concourse_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment concourse-empty \
  export-release "garden-runc/$garden_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment concourse-empty \
  export-release "bosh/$director_bosh_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment concourse-empty \
  export-release "riemann/$riemann_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment concourse-empty \
  export-release "grafana/$grafana_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment concourse-empty \
  export-release "influxdb/$influxdb_release_version" "ubuntu-trusty/$concourse_stemcell_version"

compiled_concourse_release=$(ls concourse-$concourse_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_garden_release=$(ls garden-runc-$garden_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_director_bosh_release=$(ls bosh-$director_bosh_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_riemann_release=$(ls riemann-$riemann_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_grafana_release=$(ls grafana-$grafana_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)
compiled_influxdb_release=$(ls influxdb-$influxdb_release_version-ubuntu-trusty-$concourse_stemcell_version-*.tgz)

aws s3 cp --acl public-read "$compiled_concourse_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_concourse_release"
aws s3 cp --acl public-read "$compiled_garden_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_garden_release"
aws s3 cp --acl public-read "$compiled_director_bosh_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_release"
aws s3 cp --acl public-read "$compiled_riemann_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_riemann_release"
aws s3 cp --acl public-read "$compiled_grafana_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_grafana_release"
aws s3 cp --acl public-read "$compiled_influxdb_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_influxdb_release"

aws s3 cp --acl public-read "concourse-bosh-release/fly_darwin_amd64" "s3://$PUBLIC_ARTIFACTS_BUCKET/fly_darwin_amd64-$concourse_release_version"
aws s3 cp --acl public-read "concourse-bosh-release/fly_linux_amd64" "s3://$PUBLIC_ARTIFACTS_BUCKET/fly_linux_amd64-$concourse_release_version"
aws s3 cp --acl public-read "concourse-bosh-release/fly_windows_amd64.exe" "s3://$PUBLIC_ARTIFACTS_BUCKET/fly_windows_amd64-$concourse_release_version.exe"

director_bosh_release_sha1=$(sha1sum "$compiled_director_bosh_release" | awk '{ print $1 }')
director_bosh_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_release"
concourse_release_sha1=$(sha1sum "$compiled_concourse_release" | awk '{ print $1 }')
concourse_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_concourse_release"
garden_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_garden_release"
garden_release_sha1=$(sha1sum "$compiled_garden_release" | awk '{ print $1 }')
riemann_release_sha1=$(sha1sum "$compiled_riemann_release" | awk '{ print $1 }')
riemann_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_riemann_release"
grafana_release_sha1=$(sha1sum "$compiled_grafana_release" | awk '{ print $1 }')
grafana_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_grafana_release"
influxdb_release_sha1=$(sha1sum "$compiled_influxdb_release" | awk '{ print $1 }')
influxdb_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_influxdb_release"

fly_darwin_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/fly_darwin_amd64-$concourse_release_version"
fly_linux_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/fly_linux_amd64-$concourse_release_version"
fly_windows_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/fly_windows_amd64-$concourse_release_version.exe"

director_darwin_binary_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-2.0.28-darwin-amd64"
director_linux_binary_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-2.0.28-linux-amd64"
director_windows_binary_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-2.0.28-windows-amd64.exe"

terraform_darwin_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/terraform_darwin_amd64-$concourse_release_version"
terraform_linux_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/terraform_linux_amd64-$concourse_release_version"
terraform_windows_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/terraform_windows_amd64-$concourse_release_version.exe"

wget "https://releases.hashicorp.com/terraform/0.10.2/terraform_0.10.2_darwin_amd64.zip"
unzip "terraform_0.10.2_darwin_amd64.zip"
aws s3 cp --acl public-read ./terraform "s3://$PUBLIC_ARTIFACTS_BUCKET/terraform_darwin_amd64-$concourse_release_version"
rm terraform
wget "https://releases.hashicorp.com/terraform/0.10.2/terraform_0.10.2_linux_amd64.zip"
unzip "terraform_0.10.2_linux_amd64.zip"
aws s3 cp --acl public-read ./terraform "s3://$PUBLIC_ARTIFACTS_BUCKET/terraform_linux_amd64-$concourse_release_version"
rm terraform
wget "https://releases.hashicorp.com/terraform/0.10.2/terraform_0.10.2_windows_amd64.zip"
unzip "terraform_0.10.2_windows_amd64.zip"
aws s3 cp --acl public-read ./terraform.exe "s3://$PUBLIC_ARTIFACTS_BUCKET/terraform_windows_amd64-$concourse_release_version.exe"
rm terraform.exe

echo "{
  \"concourse_stemcell_url\": \"$concourse_stemcell_url\",
  \"concourse_stemcell_sha1\": \"$concourse_stemcell_sha1\",
  \"concourse_stemcell_version\": \"$concourse_stemcell_version\",
  \"director_stemcell_url\": \"$director_stemcell_url\",
  \"director_stemcell_sha1\": \"$director_stemcell_sha1\",
  \"director_stemcell_version\": \"$director_stemcell_version\",
  \"director_bosh_release_url\": \"$director_bosh_release_url\",
  \"director_bosh_release_sha1\": \"$director_bosh_release_sha1\",
  \"director_bosh_release_version\": \"$director_bosh_release_version\",
  \"director_bosh_cpi_release_url\": \"$director_bosh_cpi_release_url\",
  \"director_bosh_cpi_release_sha1\": \"$director_bosh_cpi_release_sha1\",
  \"director_bosh_cpi_release_version\": \"$director_bosh_cpi_release_version\",
  \"concourse_release_url\": \"$concourse_release_url\",
  \"concourse_release_sha1\": \"$concourse_release_sha1\",
  \"concourse_release_version\": \"$concourse_release_version\",
  \"garden_release_url\": \"$garden_release_url\",
  \"garden_release_sha1\": \"$garden_release_sha1\",
  \"garden_release_version\": \"$garden_release_version\",
  \"riemann_release_url\": \"$riemann_release_url\",
  \"riemann_release_sha1\": \"$riemann_release_sha1\",
  \"riemann_release_version\": \"$riemann_release_version\",
  \"grafana_release_url\": \"$grafana_release_url\",
  \"grafana_release_sha1\": \"$grafana_release_sha1\",
  \"grafana_release_version\": \"$grafana_release_version\",
  \"influxdb_release_url\": \"$influxdb_release_url\",
  \"influxdb_release_sha1\": \"$influxdb_release_sha1\",
  \"influxdb_release_version\": \"$influxdb_release_version\",
  \"fly_darwin_binary_url\": \"$fly_darwin_binary_url\",
  \"fly_linux_binary_url\": \"$fly_linux_binary_url\",
  \"fly_windows_binary_url\": \"$fly_windows_binary_url\",
  \"director_darwin_binary_url\": \"$director_darwin_binary_url\",
  \"director_linux_binary_url\": \"$director_linux_binary_url\",
  \"director_windows_binary_url\": \"$director_windows_binary_url\",
  \"terraform_darwin_binary_url\": \"$terraform_darwin_binary_url\",
  \"terraform_linux_binary_url\": \"$terraform_linux_binary_url\",
  \"terraform_windows_binary_url\": \"$terraform_windows_binary_url\"
}" > compilation-vars/compilation-vars.json
