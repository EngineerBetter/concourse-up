#!/bin/bash

set -eu

version=$(cat version/version)
concourse_stemcell_url=$(cat compilation-vars/compilation-vars.json | jq -r .concourse_stemcell_url)
concourse_stemcell_sha1=$(cat compilation-vars/compilation-vars.json | jq -r .concourse_stemcell_sha1)
concourse_stemcell_version=$(cat compilation-vars/compilation-vars.json | jq -r .concourse_stemcell_version)
director_stemcell_url=$(cat compilation-vars/compilation-vars.json | jq -r .director_stemcell_url)
director_stemcell_sha1=$(cat compilation-vars/compilation-vars.json | jq -r .director_stemcell_sha1)
director_stemcell_version=$(cat compilation-vars/compilation-vars.json | jq -r .director_stemcell_version)
director_bosh_release_url=$(cat compilation-vars/compilation-vars.json | jq -r .director_bosh_release_url)
director_bosh_release_sha1=$(cat compilation-vars/compilation-vars.json | jq -r .director_bosh_release_sha1)
director_bosh_release_version=$(cat compilation-vars/compilation-vars.json | jq -r .director_bosh_release_version)
director_bosh_cpi_release_url=$(cat compilation-vars/compilation-vars.json | jq -r .director_bosh_cpi_release_url)
director_bosh_cpi_release_sha1=$(cat compilation-vars/compilation-vars.json | jq -r .director_bosh_cpi_release_sha1)
director_bosh_cpi_release_version=$(cat compilation-vars/compilation-vars.json | jq -r .director_bosh_cpi_release_version)
concourse_release_url=$(cat compilation-vars/compilation-vars.json | jq -r .concourse_release_url)
concourse_release_version=$(cat compilation-vars/compilation-vars.json | jq -r .concourse_release_version)
garden_release_url=$(cat compilation-vars/compilation-vars.json | jq -r .garden_release_url)
garden_release_version=$(cat compilation-vars/compilation-vars.json | jq -r .garden_release_version)

name="Concourse Up $version"

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

git rev-parse HEAD > release-vars/commit