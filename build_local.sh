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
director_bosh_release_version=$(cat example-compilation-vars.json | jq -r .director_bosh_release_version)
director_bosh_release_sha1=$(cat example-compilation-vars.json | jq -r .director_bosh_release_sha1)
director_bosh_cpi_release_url=$(cat example-compilation-vars.json | jq -r .director_bosh_cpi_release_url)
director_bosh_cpi_release_version=$(cat example-compilation-vars.json | jq -r .director_bosh_cpi_release_version)
director_bosh_cpi_release_sha1=$(cat example-compilation-vars.json | jq -r .director_bosh_cpi_release_sha1)
concourse_release_url=$(cat example-compilation-vars.json | jq -r .concourse_release_url)
concourse_release_version=$(cat example-compilation-vars.json | jq -r .concourse_release_version)
concourse_release_sha1=$(cat example-compilation-vars.json | jq -r .concourse_release_sha1)
garden_release_url=$(cat example-compilation-vars.json | jq -r .garden_release_url)
garden_release_version=$(cat example-compilation-vars.json | jq -r .garden_release_version)
garden_release_sha1=$(cat example-compilation-vars.json | jq -r .garden_release_sha1)

go build -ldflags "
  -X github.com/EngineerBetter/concourse-up/bosh.concourseStemcellURL=$concourse_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.concourseStemcellVersion=$concourse_stemcell_version
  -X github.com/EngineerBetter/concourse-up/bosh.concourseStemcellSHA1=$concourse_stemcell_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.concourseReleaseURL=$concourse_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.concourseReleaseVersion=$concourse_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.concourseReleaseSHA1=$concourse_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.gardenReleaseURL=$garden_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.gardenReleaseVersion=$garden_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.gardenReleaseSHA1=$garden_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.directorStemcellURL=$director_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.directorStemcellSHA1=$director_stemcell_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.directorStemcellVersion=$director_stemcell_version
  -X github.com/EngineerBetter/concourse-up/bosh.directorCPIReleaseURL=$director_bosh_cpi_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.directorCPIReleaseVersion=$director_bosh_cpi_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.directorCPIReleaseSHA1=$director_bosh_cpi_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.directorReleaseURL=$director_bosh_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.directorReleaseVersion=$director_bosh_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.directorReleaseSHA1=$director_bosh_release_sha1
  -X main.concourseUpVersion=$version
" -o concourse-up

chmod +x concourse-up

echo "$PWD/concourse-up"