#!/bin/bash

set -eu

version=dev
director_stemcell_url=$(jq -r .director_stemcell_url compilation-vars.json)
director_stemcell_sha1=$(jq -r .director_stemcell_sha1 compilation-vars.json)
director_stemcell_version=$(jq -r .director_stemcell_version compilation-vars.json)
director_bpm_release_url=$(jq -r .director_bpm_release_url compilation-vars.json)
director_bpm_release_sha1=$(jq -r .director_bpm_release_sha1 compilation-vars.json)
director_bosh_release_url=$(jq -r .director_bosh_release_url compilation-vars.json)
director_bosh_release_version=$(jq -r .director_bosh_release_version compilation-vars.json)
director_bosh_release_sha1=$(jq -r .director_bosh_release_sha1 compilation-vars.json)
director_bosh_cpi_release_url=$(jq -r .director_bosh_cpi_release_url compilation-vars.json)
director_bosh_cpi_release_version=$(jq -r .director_bosh_cpi_release_version compilation-vars.json)
director_bosh_cpi_release_sha1=$(jq -r .director_bosh_cpi_release_sha1 compilation-vars.json)
fly_darwin_binary_url=$(jq -r .fly_darwin_binary_url compilation-vars.json)
fly_linux_binary_url=$(jq -r .fly_linux_binary_url compilation-vars.json)
fly_windows_binary_url=$(jq -r .fly_windows_binary_url compilation-vars.json)
director_darwin_binary_url=$(jq -r .director_darwin_binary_url compilation-vars.json)
director_linux_binary_url=$(jq -r .director_linux_binary_url compilation-vars.json)
director_windows_binary_url=$(jq -r .director_windows_binary_url compilation-vars.json)
terraform_darwin_binary_url=$(jq -r .terraform_darwin_binary_url compilation-vars.json)
terraform_linux_binary_url=$(jq -r .terraform_linux_binary_url compilation-vars.json)
terraform_windows_binary_url=$(jq -r .terraform_windows_binary_url compilation-vars.json)

go generate github.com/EngineerBetter/concourse-up/bosh
go generate github.com/EngineerBetter/concourse-up/terraform
go build -ldflags "
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellURL=$director_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellSHA1=$director_stemcell_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorStemcellVersion=$director_stemcell_version
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorBPMReleaseURL=$director_bpm_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorBPMReleaseSHA1=$director_bpm_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseURL=$director_bosh_cpi_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseVersion=$director_bosh_cpi_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorCPIReleaseSHA1=$director_bosh_cpi_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseURL=$director_bosh_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseVersion=$director_bosh_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.DirectorReleaseSHA1=$director_bosh_release_sha1
  -X github.com/EngineerBetter/concourse-up/fly.DarwinBinaryURL=$fly_darwin_binary_url
  -X github.com/EngineerBetter/concourse-up/fly.LinuxBinaryURL=$fly_linux_binary_url
  -X github.com/EngineerBetter/concourse-up/fly.WindowsBinaryURL=$fly_windows_binary_url
  -X github.com/EngineerBetter/concourse-up/fly.ConcourseUpVersion=$version
  -X github.com/EngineerBetter/concourse-up/director.DarwinBinaryURL=$director_darwin_binary_url
  -X github.com/EngineerBetter/concourse-up/director.LinuxBinaryURL=$director_linux_binary_url
  -X github.com/EngineerBetter/concourse-up/director.WindowsBinaryURL=$director_windows_binary_url
  -X github.com/EngineerBetter/concourse-up/terraform.DarwinBinaryURL=$terraform_darwin_binary_url
  -X github.com/EngineerBetter/concourse-up/terraform.LinuxBinaryURL=$terraform_linux_binary_url
  -X github.com/EngineerBetter/concourse-up/terraform.WindowsBinaryURL=$terraform_windows_binary_url
  -X main.ConcourseUpVersion=$version
" -o concourse-up

chmod +x concourse-up

echo "$PWD/concourse-up"
