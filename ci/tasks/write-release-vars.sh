#!/bin/bash

set -eu

version=$(cat version/version)
pushd compilation-vars
  concourse_stemcell_url=$(jq -r .concourse_stemcell_url compilation-vars.json)
  concourse_stemcell_version=$(jq -r .concourse_stemcell_version compilation-vars.json)
  director_stemcell_url=$(jq -r .director_stemcell_url compilation-vars.json)
  director_stemcell_version=$(jq -r .director_stemcell_version compilation-vars.json)
  concourse_release_url=$(jq -r .concourse_release_url compilation-vars.json)
  concourse_release_version=$(jq -r .concourse_release_version compilation-vars.json)
  garden_release_url=$(jq -r .garden_release_url compilation-vars.json)
  garden_release_version=$(jq -r .garden_release_version compilation-vars.json)
  director_bosh_release_url=$(jq -r .director_bosh_release_url compilation-vars.json)
  director_bosh_release_version=$(jq -r .director_bosh_release_version compilation-vars.json)
  director_bosh_cpi_release_url=$(jq -r .director_bosh_cpi_release_url compilation-vars.json)
  director_bosh_cpi_release_version=$(jq -r .director_bosh_cpi_release_version compilation-vars.json)
  director_bpm_release_url=$(jq -r .director_bpm_release_url compilation-vars.json)
  director_bpm_release_version=$(jq -r .director_bpm_release_version compilation-vars.json)
  concourse_credhub_release_version=$(jq -r .credhub_release_version compilation-vars.json)
  concourse_credhub_release_url=$(jq -r .credhub_release_url compilation-vars.json)
  concourse_grafana_release_version=$(jq -r .grafana_release_version compilation-vars.json)
  concourse_grafana_release_url=$(jq -r .grafana_release_url compilation-vars.json)
  concourse_influxdb_release_version=$(jq -r .influxdb_release_version compilation-vars.json)
  concourse_influxdb_release_url=$(jq -r .influxdb_release_url compilation-vars.json)
  concourse_riemann_release_version=$(jq -r .riemann_release_version compilation-vars.json)
  concourse_riemann_release_url=$(jq -r .riemann_release_url compilation-vars.json)
  concourse_uaa_release_version=$(jq -r .uaa_release_version compilation-vars.json)
  concourse_uaa_release_url=$(jq -r .uaa_release_url compilation-vars.json)
  bosh_cli_version=$(jq -r .director_linux_binary_url compilation-vars.json | awk -F- '{print $5}')
  terraform_version=$(jq -r .terraform_linux_binary_url compilation-vars.json | awk -F- '{print $NF}')
popd

name="concourse-up $version"

echo "$name" > release-vars/name

cat << EOF > release-vars/body
Auto-generated release

Deploys:

- Concourse VM stemcell [bosh-aws-xen-hvm-ubuntu-trusty-go_agent $concourse_stemcell_version]($concourse_stemcell_url)
- Director stemcell [bosh-aws-xen-hvm-ubuntu-trusty-go_agent $director_stemcell_version]($director_stemcell_url)
- Concourse [$concourse_release_version]($concourse_release_url)
- Garden RunC [$garden_release_version]($garden_release_url)
- BOSH [$director_bosh_release_version]($director_bosh_release_url)
- BOSH AWS CPI [$director_bosh_cpi_release_version]($director_bosh_cpi_release_url)
- BPM [$director_bpm_release_version]($director_bpm_release_url)
- Credhub [$concourse_credhub_release_version]($concourse_credhub_release_url)
- Grafana [$concourse_grafana_release_version]($concourse_grafana_release_url)
- InfluxDB [$concourse_influxdb_release_version]($concourse_influxdb_release_url)
- Riemann [$concourse_riemann_release_version]($concourse_riemann_release_url)
- UAA [$concourse_uaa_release_version]($concourse_uaa_release_url)
- BOSH CLI $bosh_cli_version
- Terraform $terraform_version
EOF

pushd concourse-up
  commit=$(git rev-parse HEAD)
popd

echo "$commit" > release-vars/commit
