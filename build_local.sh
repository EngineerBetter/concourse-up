#!/bin/bash

set -eu

version=dev
concourse_stemcell_url=$(jq -r .concourse_stemcell_url compilation-vars.json)
concourse_stemcell_sha1=$(jq -r .concourse_stemcell_sha1 compilation-vars.json)
concourse_stemcell_version=$(jq -r .concourse_stemcell_version compilation-vars.json)
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
concourse_release_url=$(jq -r .concourse_release_url compilation-vars.json)
concourse_release_version=$(jq -r .concourse_release_version compilation-vars.json)
concourse_release_sha1=$(jq -r .concourse_release_sha1 compilation-vars.json)
riemann_release_url=$(jq -r .riemann_release_url compilation-vars.json)
riemann_release_version=$(jq -r .riemann_release_version compilation-vars.json)
riemann_release_sha1=$(jq -r .riemann_release_sha1 compilation-vars.json)
grafana_release_url=$(jq -r .grafana_release_url compilation-vars.json)
grafana_release_version=$(jq -r .grafana_release_version compilation-vars.json)
grafana_release_sha1=$(jq -r .grafana_release_sha1 compilation-vars.json)
influxdb_release_url=$(jq -r .influxdb_release_url compilation-vars.json)
influxdb_release_version=$(jq -r .influxdb_release_version compilation-vars.json)
influxdb_release_sha1=$(jq -r .influxdb_release_sha1 compilation-vars.json)
credhub_release_url=$(jq -r .credhub_release_url compilation-vars.json)
credhub_release_version=$(jq -r .credhub_release_version compilation-vars.json)
credhub_release_sha1=$(jq -r .credhub_release_sha1 compilation-vars.json)
uaa_release_url=$(jq -r .uaa_release_url compilation-vars.json)
uaa_release_version=$(jq -r .uaa_release_version compilation-vars.json)
uaa_release_sha1=$(jq -r .uaa_release_sha1 compilation-vars.json)
garden_release_url=$(jq -r .garden_release_url compilation-vars.json)
garden_release_version=$(jq -r .garden_release_version compilation-vars.json)
garden_release_sha1=$(jq -r .garden_release_sha1 compilation-vars.json)
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
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellURL=$concourse_stemcell_url
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellVersion=$concourse_stemcell_version
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseStemcellSHA1=$concourse_stemcell_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseURL=$concourse_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseVersion=$concourse_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.ConcourseReleaseSHA1=$concourse_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.RiemannReleaseURL=$riemann_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.RiemannReleaseVersion=$riemann_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.RiemannReleaseSHA1=$riemann_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.GrafanaReleaseURL=$grafana_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.GrafanaReleaseVersion=$grafana_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.GrafanaReleaseSHA1=$grafana_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.InfluxDBReleaseURL=$influxdb_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.InfluxDBReleaseVersion=$influxdb_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.InfluxDBReleaseSHA1=$influxdb_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.CredhubReleaseURL=$credhub_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.CredhubReleaseVersion=$credhub_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.CredhubReleaseSHA1=$credhub_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.UAAReleaseURL=$uaa_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.UAAReleaseVersion=$uaa_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.UAAReleaseSHA1=$uaa_release_sha1
  -X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseURL=$garden_release_url
  -X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseVersion=$garden_release_version
  -X github.com/EngineerBetter/concourse-up/bosh.GardenReleaseSHA1=$garden_release_sha1
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
