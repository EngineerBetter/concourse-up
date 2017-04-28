#!/bin/bash

set -eu

version=dev
concourse_stemcell_url=$(cat example-compilation-vars.json | jq -r .concourse_stemcell_url)
concourse_stemcell_sha1=$(cat example-compilation-vars.json | jq -r .concourse_stemcell_sha1)
concourse_stemcell_version=$(cat example-compilation-vars.json | jq -r .concourse_stemcell_version)
director_stemcell_url=$(cat example-compilation-vars.json | jq -r .director_stemcell_url)
director_stemcell_sha1=$(cat example-compilation-vars.json | jq -r .director_stemcell_sha1)
director_stemcell_version=$(cat example-compilation-vars.json | jq -r .director_stemcell_version)
director_bosh_release_url=$(cat example-compilation-vars.json | jq -r .director_bosh_release_url)
director_bosh_release_sha1=$(cat example-compilation-vars.json | jq -r .director_bosh_release_sha1)
director_bosh_release_version=$(cat example-compilation-vars.json | jq -r .director_bosh_release_version)
director_bosh_cpi_release_url=$(cat example-compilation-vars.json | jq -r .director_bosh_cpi_release_url)
director_bosh_cpi_release_sha1=$(cat example-compilation-vars.json | jq -r .director_bosh_cpi_release_sha1)
director_bosh_cpi_release_version=$(cat example-compilation-vars.json | jq -r .director_bosh_cpi_release_version)
concourse_release_url=$(cat example-compilation-vars.json | jq -r .concourse_release_url)
concourse_release_version=$(cat example-compilation-vars.json | jq -r .concourse_release_version)
garden_release_url=$(cat example-compilation-vars.json | jq -r .garden_release_url)
garden_release_version=$(cat example-compilation-vars.json | jq -r .garden_release_version)

go build -ldflags "
  -X github.com/engineerbetter/concourse-up/bosh.concourseStemcellURL=$concourse_stemcell_url
  -X github.com/engineerbetter/concourse-up/bosh.concourseStemcellVersion=$concourse_stemcell_version
  -X github.com/engineerbetter/concourse-up/bosh.concourseCompiledReleaseURL=$concourse_release_url
  -X github.com/engineerbetter/concourse-up/bosh.concourseReleaseVersion=$concourse_release_version
  -X github.com/engineerbetter/concourse-up/bosh.gardenCompiledReleaseURL=$garden_release_url
  -X github.com/engineerbetter/concourse-up/bosh.gardenReleaseVersion=$garden_release_version
  -X github.com/engineerbetter/concourse-up/bosh.DirectorStemcellURL=$director_stemcell_url
  -X github.com/engineerbetter/concourse-up/bosh.DirectorStemcellSHA1=$director_stemcell_sha1
  -X github.com/engineerbetter/concourse-up/bosh.DirectorCPIReleaseURL=$director_bosh_cpi_release_url
  -X github.com/engineerbetter/concourse-up/bosh.DirectorCPIReleaseSHA1=$director_bosh_cpi_release_sha1
  -X github.com/engineerbetter/concourse-up/bosh.DirectorReleaseURL=$director_bosh_release_url
  -X github.com/engineerbetter/concourse-up/bosh.DirectorReleaseSHA1=$director_bosh_release_sha1
  -X main.concourseUpVersion=$version
" -o concourse-up

chmod +x concourse-up

echo "$PWD/concourse-up"