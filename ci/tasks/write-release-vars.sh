#!/bin/bash

set -eu

version=$(cat version/version)
pushd concourse-up/resources
  bin_bosh_cli_version=$(                jq -r '."bosh-cli".linux' director-versions.json)
  bin_terraform_version=$(               jq -r '.terraform.linux' director-versions.json)
  deployment_concourse_release_url=$(    jq -r '.[] | select(.value.name? == "concourse") | .value.url' versions.json)
  deployment_concourse_release_version=$(jq -r '.[] | select(.value.name? == "concourse") | .value.version' versions.json)
  deployment_credhub_release_url=$(      jq -r '.[] | select(.value.name? == "credhub") | .value.url' versions.json)
  deployment_credhub_release_version=$(  jq -r '.[] | select(.value.name? == "credhub") | .value.version' versions.json)
  deployment_garden_release_url=$(       jq -r '.[] | select(.value.name? == "garden-runc") | .value.url' versions.json)
  deployment_garden_release_version=$(   jq -r '.[] | select(.value.name? == "garden-runc") | .value.version' versions.json)
  deployment_grafana_release_url=$(      jq -r '.[] | select(.value.name? == "grafana") | .value.url' versions.json)
  deployment_grafana_release_version=$(  jq -r '.[] | select(.value.name? == "grafana") | .value.version' versions.json)
  deployment_influxdb_release_url=$(     jq -r '.[] | select(.value.name? == "influxdb") | .value.url' versions.json)
  deployment_influxdb_release_version=$( jq -r '.[] | select(.value.name? == "influxdb") | .value.version' versions.json)
  deployment_riemann_release_url=$(      jq -r '.[] | select(.value.name? == "riemann") | .value.url' versions.json)
  deployment_riemann_release_version=$(  jq -r '.[] | select(.value.name? == "riemann") | .value.version' versions.json)
  deployment_stemcell_version=$(         jq -r '.[] | select(.path == "/stemcells/alias=xenial/version") | .value' versions.json)
  deployment_uaa_release_url=$(          jq -r '.[] | select(.value.name? == "uaa") | .value.url' versions.json)
  deployment_uaa_release_version=$(      jq -r '.[] | select(.value.name? == "uaa") | .value.version' versions.json)
  director_bosh_cpi_release_url=$(       jq -r .cpi.url director-versions.json)
  director_bosh_cpi_release_version=$(   jq -r .cpi.version director-versions.json)
  director_bosh_release_url=$(           jq -r .bosh.url director-versions.json)
  director_bosh_release_version=$(       jq -r .bosh.version director-versions.json)
  director_bpm_release_url=$(            jq -r .bpm.url director-versions.json)
  director_bpm_release_version=$(        jq -r .bpm.version director-versions.json)
  director_stemcell_version=$(           jq -r .stemcell.url director-versions.json | cut -d= -f2)
popd

name="concourse-up $version"

echo "$name" > release-vars/name

# TODO: remove notice once github auth is supported again
cat << EOF > release-vars/body
Concourse 4.0 has changed the way that Authentication options are configured, which must now be set at deploy time. If you previously used \`fly\` to configure third party authentication providers (e.g GitHub oAuth) then this option is no longer available. The Concourse-Up team are currently looking at solutions to this problem
---

Auto-generated release

Deploys:

- Concourse VM stemcell bosh-aws-xen-hvm-ubuntu-xenial-go_agent $deployment_stemcell_version
- Director stemcell     bosh-aws-xen-hvm-ubuntu-xenial-go_agent $director_stemcell_version
- Concourse [$deployment_concourse_release_version]($deployment_concourse_release_url)
- Garden RunC [$deployment_garden_release_version]($deployment_garden_release_url)
- BOSH [$director_bosh_release_version]($director_bosh_release_url)
- BOSH AWS CPI [$director_bosh_cpi_release_version]($director_bosh_cpi_release_url)
- BPM [$director_bpm_release_version]($director_bpm_release_url)
- Credhub [$deployment_credhub_release_version]($deployment_credhub_release_url)
- Grafana [$deployment_grafana_release_version]($deployment_grafana_release_url)
- InfluxDB [$deployment_influxdb_release_version]($deployment_influxdb_release_url)
- Riemann [$deployment_riemann_release_version]($deployment_riemann_release_url)
- UAA [$deployment_uaa_release_version]($deployment_uaa_release_url)
- BOSH CLI $bin_bosh_cli_version
- Terraform $bin_terraform_version
EOF

pushd concourse-up
  commit=$(git rev-parse HEAD)
popd

echo "$commit" > release-vars/commit
