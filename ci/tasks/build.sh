#!/bin/bash

set -eu

build_dir=$PWD/build
mkdir -p build_dir

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

mkdir -p $GOPATH/src/github.com/EngineerBetter/concourse-up
mv concourse-up/* $GOPATH/src/github.com/EngineerBetter/concourse-up
cd $GOPATH/src/github.com/EngineerBetter/concourse-up

go build -ldflags "
  -X main.concourseUpVersion=$version
  -X github.com/EngineerBetter/concourse-up/bosh.concourseStemcellURL=$concourse_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.concourseStemcellVersion=$concourse_stemcell_version
  -X github.com/EngineerBetter/concourse-up/bosh.concourseCompiledReleaseURL=$concourse_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.concourseReleaseVersion=$concourse_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.gardenCompiledReleaseURL=$garden_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.gardenReleaseVersion=$garden_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.directorStemcellURL=$director_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.directorStemcellSHA1=$director_stemcell_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.directorCPIReleaseURL=$director_bosh_cpi_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.directorCPIReleaseSHA1=$director_bosh_cpi_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.directorReleaseURL=$director_bosh_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.directorReleaseSHA1=$director_bosh_release_sha1
" -o $build_dir/$OUTPUT_FILE
