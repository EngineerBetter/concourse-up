#!/bin/bash
# shellcheck disable=SC2091,SC2006

# Disabling SC2091 above because we want to print commands encased in $()
# Disabling SC2006 above because ``` code blocks are misinterpretted as shell execution

set -eu

version=$(cat version/version)
pushd concourse-up-ops
  bin_bosh_cli_version=$(                jq -r '."bosh-cli".linux' director-versions.json)
  bin_terraform_version=$(               jq -r '.terraform.linux' director-versions.json)
  deployment_concourse_release_url=$(    jq -r '.[] | select(.value.name? == "concourse") | .value.url' ops/versions.json)
  deployment_concourse_release_version=$(jq -r '.[] | select(.value.name? == "concourse") | .value.version' ops/versions.json)
  deployment_credhub_release_url=$(      jq -r '.[] | select(.value.name? == "credhub") | .value.url' ops/versions.json)
  deployment_credhub_release_version=$(  jq -r '.[] | select(.value.name? == "credhub") | .value.version' ops/versions.json)
  deployment_garden_release_url=$(       jq -r '.[] | select(.value.name? == "garden-runc") | .value.url' ops/versions.json)
  deployment_garden_release_version=$(   jq -r '.[] | select(.value.name? == "garden-runc") | .value.version' ops/versions.json)
  deployment_grafana_release_url=$(      jq -r '.[] | select(.value.name? == "grafana") | .value.url' ops/versions.json)
  deployment_grafana_release_version=$(  jq -r '.[] | select(.value.name? == "grafana") | .value.version' ops/versions.json)
  deployment_influxdb_release_url=$(     jq -r '.[] | select(.value.name? == "influxdb") | .value.url' ops/versions.json)
  deployment_influxdb_release_version=$( jq -r '.[] | select(.value.name? == "influxdb") | .value.version' ops/versions.json)
  deployment_riemann_release_url=$(      jq -r '.[] | select(.value.name? == "riemann") | .value.url' ops/versions.json)
  deployment_riemann_release_version=$(  jq -r '.[] | select(.value.name? == "riemann") | .value.version' ops/versions.json)
  deployment_stemcell_version=$(         jq -r '.[] | select(.path == "/stemcells/alias=xenial/version") | .value' ops/versions.json)
  deployment_uaa_release_url=$(          jq -r '.[] | select(.value.name? == "uaa") | .value.url' ops/versions.json)
  deployment_uaa_release_version=$(      jq -r '.[] | select(.value.name? == "uaa") | .value.version' ops/versions.json)
  director_bosh_cpi_release_url=$(       jq -r .cpi.url director-versions.json)
  director_bosh_cpi_release_version=$(   jq -r .cpi.version director-versions.json)
  director_bosh_release_url=$(           jq -r .bosh.url director-versions.json)
  director_bosh_release_version=$(       jq -r .bosh.version director-versions.json)
  director_bpm_release_url=$(            jq -r .bpm.url director-versions.json)
  director_bpm_release_version=$(        jq -r .bpm.version director-versions.json)
  director_stemcell_version=$(           jq -r .stemcell.url director-versions.json | cut -d= -f2)
popd

pushd ops-version
  ops_version=$(cat version)
popd

name="concourse-up $version"

echo "$name" > release-vars/name

cat << EOF > release-vars/body

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

>Note to build locally you will need to clone [concourse-up-ops](https://github.com/EngineerBetter/concourse-up-ops/tree/$ops_version) (version $ops_version) to the same level as concourse-up to get the required manifests and ops files.
EOF

cat <<EOF > release-vars/slackmsg
<!channel> Concourse Up $(cat version/version) published to Github
```
$(diff --suppress-common-lines release-vars/body <(curl -Ss https://api.github.com/repos/EngineerBetter/concourse-up/releases/latest | jq -r .body) || true)
```
EOF

pushd concourse-up
  commit=$(git rev-parse HEAD)
popd

echo "$commit" > release-vars/commit
