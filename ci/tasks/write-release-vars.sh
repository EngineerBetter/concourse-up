#!/bin/bash

set -eu

version=$(cat version/version)
pushd compilation-vars
  concourse_stemcell_url=$(jq -r .concourse_stemcell_url | compilation-vars.json)
  concourse_stemcell_version=$(jq -r .concourse_stemcell_version | compilation-vars.json)
  director_stemcell_url=$(jq -r .director_stemcell_url | compilation-vars.json)
  director_stemcell_version=$(jq -r .director_stemcell_version | compilation-vars.json)
  director_bosh_release_url=$(jq -r .director_bosh_release_url | compilation-vars.json)
  director_bosh_release_version=$(jq -r .director_bosh_release_version | compilation-vars.json)
  director_bosh_cpi_release_url=$(jq -r .director_bosh_cpi_release_url | compilation-vars.json)
  director_bosh_cpi_release_version=$(jq -r .director_bosh_cpi_release_version | compilation-vars.json)
  concourse_release_url=$(jq -r .concourse_release_url | compilation-vars.json)
  concourse_release_version=$(jq -r .concourse_release_version | compilation-vars.json)
  garden_release_url=$(jq -r .garden_release_url | compilation-vars.json)
  garden_release_version=$(jq -r .garden_release_version | compilation-vars.json)
popd

name="concourse-up $version"

echo "$name" > release-vars/name

body="Auto-generated release

Deploys:
- BOSH [$director_bosh_release_version]($director_bosh_release_url)
- BOSH AWS CPI [$director_bosh_cpi_release_version]($director_bosh_cpi_release_url)
- Director stemcell [bosh-aws-xen-hvm-ubuntu-trusty-go_agent $director_stemcell_version]($director_stemcell_url)
- Concourse [$concourse_release_version]($concourse_release_url)
- Garden RunC [$garden_release_version]($garden_release_url)
- Concourse VM stemcell [bosh-aws-xen-hvm-ubuntu-trusty-go_agent $concourse_stemcell_version]($concourse_stemcell_url)"

echo "$body" > release-vars/body

pushd concourse-up
  commit=$(git rev-parse HEAD)
popd

echo "$commit" > release-vars/commit
