#!/bin/bash

set -eu
set -x

# delete any concourse versions that have pre-compiled packages
rm -f concourse-github-release/concourse-*-*.tgz

echo "$BOSH_CA_CERT" > bosh_ca_cert.pem

bosh="bosh --non-interactive --environment $BOSH_TARGET --client $BOSH_USERNAME --client-secret $BOSH_PASSWORD --ca-cert bosh_ca_cert.pem"

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

uaa_release_version=$(cat uaa-release/version)
uaa_release_url=$(cat uaa-release/url)
uaa_release_sha1=$(cat uaa-release/sha1)

credhub_release_version=$(cat credhub-release/version)
credhub_release_url=$(cat credhub-release/url)
credhub_release_sha1=$(cat credhub-release/sha1)

director_bosh_release_version=$(bosh int bosh-deployment/bosh.yml --path /releases/name=bosh/version)
director_bosh_release_url=$(bosh int bosh-deployment/bosh.yml --path /releases/name=bosh/url)

director_bpm_release_version=$(bosh int bosh-deployment/bosh.yml --path /releases/name=bpm/version)
director_bpm_release_url=$(bosh int bosh-deployment/bosh.yml --path /releases/name=bpm/url)

concourse_release_version=$(cat concourse-release/version)
garden_release_version=$(cat garden-runc-release/version)

$bosh upload-stemcell "concourse-stemcell/stemcell.tgz"
$bosh upload-release "garden-runc-release/release.tgz"
$bosh upload-release "concourse-release/release.tgz"
$bosh upload-release "$director_bosh_release_url"
$bosh upload-release "$director_bpm_release_url"
$bosh upload-release "director-bosh-cpi-release/release.tgz"
$bosh upload-release "riemann-release/release.tgz"
$bosh upload-release "grafana-release/release.tgz"
$bosh upload-release "influxdb-release/release.tgz"
$bosh upload-release "uaa-release/release.tgz"
$bosh upload-release "credhub-release/release.tgz"

echo "---
name: cup-compilation-workspace

releases:
- name: concourse
  version: \"$concourse_release_version\"
- name: garden-runc
  version: \"$garden_release_version\"
- name: bosh
  version: \"$director_bosh_release_version\"
- name: bpm
  version: \"$director_bpm_release_version\"
- name: bosh-aws-cpi
  version: \"$director_bosh_cpi_release_version\"
- name: riemann
  version: \"$riemann_release_version\"
- name: grafana
  version: \"$grafana_release_version\"
- name: influxdb
  version: \"$influxdb_release_version\"
- name: uaa
  version: \"$uaa_release_version\"
- name: credhub
  version: \"$credhub_release_version\"

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: \"$concourse_stemcell_version\"

instance_groups: []

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000" > cup-compilation-workspace.yml

$bosh \
  --deployment cup-compilation-workspace \
  deploy \
  cup-compilation-workspace.yml

# avoids compiling Windows jobs released in Concourse 3.14
$bosh \
  --deployment cup-compilation-workspace \
  export-release "concourse/$concourse_release_version" "ubuntu-trusty/$concourse_stemcell_version" \
  --job={atc,baggageclaim,bbr-atcdb,blackbox,tsa,worker}

$bosh \
  --deployment cup-compilation-workspace \
  export-release "garden-runc/$garden_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment cup-compilation-workspace \
  export-release "bosh/$director_bosh_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment cup-compilation-workspace \
  export-release "bpm/$director_bpm_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment cup-compilation-workspace \
  export-release "riemann/$riemann_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment cup-compilation-workspace \
  export-release "grafana/$grafana_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment cup-compilation-workspace \
  export-release "influxdb/$influxdb_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment cup-compilation-workspace \
  export-release "uaa/$uaa_release_version" "ubuntu-trusty/$concourse_stemcell_version"

$bosh \
  --deployment cup-compilation-workspace \
  export-release "credhub/$credhub_release_version" "ubuntu-trusty/$concourse_stemcell_version"

compiled_concourse_release=$(echo concourse-"$concourse_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_garden_release=$(echo garden-runc-"$garden_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_director_bosh_release=$(echo bosh-"$director_bosh_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_director_bpm_release=$(echo bpm-"$director_bpm_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_riemann_release=$(echo riemann-"$riemann_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_grafana_release=$(echo grafana-"$grafana_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_influxdb_release=$(echo influxdb-"$influxdb_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_uaa_release=$(echo uaa-"$uaa_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)
compiled_credhub_release=$(echo credhub-"$credhub_release_version"-ubuntu-trusty-"$concourse_stemcell_version"-*.tgz)

aws s3 cp --acl public-read "$compiled_concourse_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_concourse_release"
aws s3 cp --acl public-read "$compiled_garden_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_garden_release"
aws s3 cp --acl public-read "$compiled_director_bosh_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_release"
aws s3 cp --acl public-read "$compiled_director_bpm_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bpm_release"
aws s3 cp --acl public-read "$compiled_riemann_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_riemann_release"
aws s3 cp --acl public-read "$compiled_grafana_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_grafana_release"
aws s3 cp --acl public-read "$compiled_influxdb_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_influxdb_release"
aws s3 cp --acl public-read "$compiled_uaa_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_uaa_release"
aws s3 cp --acl public-read "$compiled_credhub_release" "s3://$PUBLIC_ARTIFACTS_BUCKET/$compiled_credhub_release"

aws s3 cp --acl public-read "concourse-github-release/fly_darwin_amd64" "s3://$PUBLIC_ARTIFACTS_BUCKET/fly_darwin_amd64-$concourse_release_version"
aws s3 cp --acl public-read "concourse-github-release/fly_linux_amd64" "s3://$PUBLIC_ARTIFACTS_BUCKET/fly_linux_amd64-$concourse_release_version"
aws s3 cp --acl public-read "concourse-github-release/fly_windows_amd64.exe" "s3://$PUBLIC_ARTIFACTS_BUCKET/fly_windows_amd64-$concourse_release_version.exe"

director_bosh_release_sha1=$(sha1sum "$compiled_director_bosh_release" | awk '{ print $1 }')
director_bosh_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bosh_release"
director_bpm_release_sha1=$(sha1sum "$compiled_director_bpm_release" | awk '{ print $1 }')
director_bpm_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_director_bpm_release"
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
uaa_release_sha1=$(sha1sum "$compiled_uaa_release" | awk '{ print $1 }')
uaa_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_uaa_release"
credhub_release_sha1=$(sha1sum "$compiled_credhub_release" | awk '{ print $1 }')
credhub_release_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/$compiled_credhub_release"

fly_darwin_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/fly_darwin_amd64-$concourse_release_version"
fly_linux_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/fly_linux_amd64-$concourse_release_version"
fly_windows_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/fly_windows_amd64-$concourse_release_version.exe"

bosh_cli_version="2.0.40"
director_darwin_binary_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-${bosh_cli_version}-darwin-amd64"
director_linux_binary_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-${bosh_cli_version}-linux-amd64"
director_windows_binary_url="https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-${bosh_cli_version}-windows-amd64.exe"

terraform_version="0.11.4"

wget "https://releases.hashicorp.com/terraform/${terraform_version}/terraform_${terraform_version}_darwin_amd64.zip"
unzip "terraform_${terraform_version}_darwin_amd64.zip"
aws s3 cp --acl public-read ./terraform "s3://$PUBLIC_ARTIFACTS_BUCKET/terraform_darwin_amd64-${terraform_version}"
rm terraform
wget "https://releases.hashicorp.com/terraform/${terraform_version}/terraform_${terraform_version}_linux_amd64.zip"
unzip "terraform_${terraform_version}_linux_amd64.zip"
aws s3 cp --acl public-read ./terraform "s3://$PUBLIC_ARTIFACTS_BUCKET/terraform_linux_amd64-${terraform_version}"
rm terraform
wget "https://releases.hashicorp.com/terraform/${terraform_version}/terraform_${terraform_version}_windows_amd64.zip"
unzip "terraform_${terraform_version}_windows_amd64.zip"
aws s3 cp --acl public-read ./terraform.exe "s3://$PUBLIC_ARTIFACTS_BUCKET/terraform_windows_amd64-${terraform_version}.exe"
rm terraform.exe

terraform_darwin_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/terraform_darwin_amd64-$terraform_version"
terraform_linux_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/terraform_linux_amd64-$terraform_version"
terraform_windows_binary_url="https://s3-$AWS_DEFAULT_REGION.amazonaws.com/$PUBLIC_ARTIFACTS_BUCKET/terraform_windows_amd64-$terraform_version.exe"

echo "{
  \"concourse_stemcell_url\": \"$concourse_stemcell_url\",
  \"concourse_stemcell_sha1\": \"$concourse_stemcell_sha1\",
  \"concourse_stemcell_version\": \"$concourse_stemcell_version\",
  \"director_stemcell_url\": \"$director_stemcell_url\",
  \"director_stemcell_sha1\": \"$director_stemcell_sha1\",
  \"director_stemcell_version\": \"$director_stemcell_version\",
  \"director_bpm_release_url\": \"$director_bpm_release_url\",
  \"director_bpm_release_sha1\": \"$director_bpm_release_sha1\",
  \"director_bpm_release_version\": \"$director_bpm_release_version\",
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
  \"uaa_release_url\": \"$uaa_release_url\",
  \"uaa_release_sha1\": \"$uaa_release_sha1\",
  \"uaa_release_version\": \"$uaa_release_version\",
  \"credhub_release_url\": \"$credhub_release_url\",
  \"credhub_release_sha1\": \"$credhub_release_sha1\",
  \"credhub_release_version\": \"$credhub_release_version\",
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
