#!/bin/bash

set -eu

version=dev
concourse_stemcell_url=$(jq -r .concourse_stemcell_url compilation-vars.json)
concourse_stemcell_sha1=$(jq -r .concourse_stemcell_sha1 compilation-vars.json)
concourse_stemcell_version=$(jq -r .concourse_stemcell_version compilation-vars.json)
director_stemcell_url=$(jq -r .director_stemcell_url compilation-vars.json)
director_stemcell_sha1=$(jq -r .director_stemcell_sha1 compilation-vars.json)
director_stemcell_version=$(jq -r .director_stemcell_version compilation-vars.json)
director_bosh_release_url=$(jq -r .director_bosh_release_url compilation-vars.json)
director_bosh_release_version=$(jq -r .director_bosh_release_version compilation-vars.json)
director_bosh_release_sha1=$(jq -r .director_bosh_release_sha1 compilation-vars.json)
director_bosh_cpi_release_url=$(jq -r .director_bosh_cpi_release_url compilation-vars.json)
director_bosh_cpi_release_version=$(jq -r .director_bosh_cpi_release_version compilation-vars.json)
director_bosh_cpi_release_sha1=$(jq -r .director_bosh_cpi_release_sha1 compilation-vars.json)
concourse_release_url=$(jq -r .concourse_release_url compilation-vars.json)
concourse_release_version=$(jq -r .concourse_release_version compilation-vars.json)
concourse_release_sha1=$(jq -r .concourse_release_sha1 compilation-vars.json)
garden_release_url=$(jq -r .garden_release_url compilation-vars.json)
garden_release_version=$(jq -r .garden_release_version compilation-vars.json)
garden_release_sha1=$(jq -r .garden_release_sha1 compilation-vars.json)
darwin_binary_url=$(jq -r .darwin_binary_url compilation-vars.json)
linux_binary_url=$(jq -r .linux_binary_url compilation-vars.json)
windows_binary_url=$(jq -r .windows_binary_url compilation-vars.json)

go build -ldflags "
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellURL=$concourse_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellVersion=$concourse_stemcell_version
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellSHA1=$concourse_stemcell_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseURL=$concourse_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseVersion=$concourse_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseSHA1=$concourse_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseURL=$garden_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseVersion=$garden_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseSHA1=$garden_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellURL=$director_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellSHA1=$director_stemcell_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellVersion=$director_stemcell_version
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseURL=$director_bosh_cpi_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseVersion=$director_bosh_cpi_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseSHA1=$director_bosh_cpi_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseURL=$director_bosh_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseVersion=$director_bosh_release_version
  -X github.com/EngineerBetter/concourse-up/fly.DarwinBinaryURL=$darwin_binary_url
  -X github.com/EngineerBetter/concourse-up/fly.LinuxBinaryURL=$linux_binary_url
  -X github.com/EngineerBetter/concourse-up/fly.WindowsBinaryURL=$windows_binary_url
  -X main.ConcourseUpVersion=$version
" -o concourse-up

chmod +x concourse-up

echo "$PWD/concourse-up"