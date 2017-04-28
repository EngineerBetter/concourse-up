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

mkdir -p $GOPATH/src/github.com/engineerbetter/concourse-up
mv concourse-up/* $GOPATH/src/github.com/engineerbetter/concourse-up
cd $GOPATH/src/github.com/engineerbetter/concourse-up

go build -ldflags "
  -X main.concourseUpVersion=$version

  -X bosh.concourseStemcellURL=$concourse_stemcell_url
  -X bosh.concourseCompiledReleaseURL=$concourse_release_url
  -X bosh.concourseReleaseVersion=$concourse_release_version
  -X bosh.gardenCompiledReleaseURL=$garden_release_url
  -X bosh.gardenReleaseVersion=$garden_release_version

  -X bosh.directorStemcellURL=$director_stemcell_url
  -X bosh.directorStemcellSHA1=$director_stemcell_sha1
  -X bosh.directorCPIReleaseURL=$director_bosh_cpi_release_url
  -X bosh.directorCPIReleaseSHA1=$director_bosh_cpi_release_sha1
  -X bosh.directorReleaseURL=$director_bosh_release_sha1
  -X bosh.directorReleaseSHA1=$director_bosh_release_version
" -o $build_dir/$OUTPUT_FILE
