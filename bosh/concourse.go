package bosh

import (
	"fmt"

	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/db"
	"github.com/engineerbetter/concourse-up/terraform"
	"github.com/engineerbetter/concourse-up/util"
)

const concourseManifestFilename = "concourse.yml"
const concourseDeploymentName = "concourse"

var concourseStemcellURL = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellURL"
var concourseStemcellVersion = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion"
var concourseCompiledReleaseURL = "COMPILE_TIME_VARIABLE_bosh_concourseCompiledReleaseURL"
var concourseReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseVersion"
var gardenCompiledReleaseURL = "COMPILE_TIME_VARIABLE_bosh_gardenCompiledReleaseURL"
var gardenReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseVersion"

func (client *Client) uploadConcourse() error {
	_, err := client.director.RunAuthenticatedCommand(
		"upload-stemcell",
		concourseStemcellURL,
	)
	if err != nil {
		return err
	}

	for _, releaseURL := range []string{concourseCompiledReleaseURL, gardenCompiledReleaseURL} {
		_, err = client.director.RunAuthenticatedCommand(
			"upload-release",
			releaseURL,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) deployConcourse() error {
	concourseManifestBytes, err := generateConcourseManifest(client.config, client.metadata)
	if err != nil {
		return err
	}

	concourseManifestPath, err := client.director.SaveFileToWorkingDir(concourseManifestFilename, concourseManifestBytes)
	if err != nil {
		return err
	}

	_, err = client.director.RunAuthenticatedCommand(
		"--deployment",
		concourseDeploymentName,
		"deploy",
		concourseManifestPath,
	)
	if err != nil {
		return err
	}

	return nil
}

func generateConcourseManifest(config *config.Config, metadata *terraform.Metadata) ([]byte, error) {
	templateParams := awsConcourseManifestParams{
		WorkerCount:             config.ConcourseWorkerCount,
		WorkerSize:              config.ConcourseWorkerSize,
		URL:                     fmt.Sprintf("https://%s", config.Domain),
		Username:                config.ConcourseUsername,
		Password:                config.ConcoursePassword,
		DBUsername:              config.RDSUsername,
		DBPassword:              config.RDSPassword,
		DBName:                  config.ConcourseDBName,
		DBHost:                  metadata.BoshDBAddress.Value,
		DBPort:                  metadata.BoshDBPort.Value,
		DBCACert:                db.RDSRootCert,
		Project:                 config.Project,
		Network:                 "default",
		TLSCert:                 config.ConcourseCert,
		TLSKey:                  config.ConcourseKey,
		AllowSelfSignedCerts:    "true",
		ConcourseReleaseVersion: concourseReleaseVersion,
		GardenReleaseVersion:    gardenReleaseVersion,
		StemcellVersion:         concourseStemcellVersion,
	}

	return util.RenderTemplate(awsConcourseManifestTemplate, templateParams)
}

type awsConcourseManifestParams struct {
	WorkerCount             int
	WorkerSize              string
	URL                     string
	Username                string
	Password                string
	DBHost                  string
	DBName                  string
	DBPort                  string
	DBUsername              string
	DBPassword              string
	Project                 string
	DBCACert                string
	Network                 string
	TLSCert                 string
	TLSKey                  string
	AllowSelfSignedCerts    string
	ConcourseReleaseVersion string
	GardenReleaseVersion    string
	StemcellVersion         string
}

// Indent is a helper function to indent the field a given number of spaces
func (params awsConcourseManifestParams) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

const awsConcourseManifestTemplate = `---
name: concourse

releases:
- name: concourse
  version: <% .ConcourseReleaseVersion %>
- name: garden-runc
  version: <% .GardenReleaseVersion %>

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: <% .StemcellVersion %>

tags:
  concourse-up-project: <% .Project %>
  concourse-up-component: concourse

instance_groups:
- name: web
  instances: 1
  vm_type: concourse-medium
  stemcell: trusty
  azs:
  - z1
  networks:
  - name: <% .Network %>
    default: [dns, gateway]
  vm_extensions:
  - elb
  jobs:
  - name: atc
    release: concourse
    properties:
      allow_self_signed_certificates: <% .AllowSelfSignedCerts %>
      external_url: <% .URL %>
      basic_auth_username: <% .Username %>
      basic_auth_password: <% .Password %>
      tls_cert: |-
        <% .Indent "8" .TLSCert %>

      tls_key: |-
        <% .Indent "8" .TLSKey %>

      postgresql:
        port: <% .DBPort %>
        database: <% .DBName %>
        role:
          name: <% .DBUsername %>
          password:  <% .DBPassword %>
        host: <% .DBHost %>
        ssl_mode: verify-full
        ca_cert: |-
          <% .Indent "10" .DBCACert %>

  - name: tsa
    release: concourse
    properties: {}

- name: worker
  instances: <% .WorkerCount %>
  vm_type: concourse-<% .WorkerSize %>
  stemcell: trusty
  azs:
  - z1
  networks:
  - name: default
    default: [dns, gateway]
  jobs:
  - name: groundcrew
    release: concourse
    properties: {}
  - name: baggageclaim
    release: concourse
    properties: {}
  - name: garden
    release: garden-runc
    properties:
      garden:
        listen_network: tcp
        listen_address: 0.0.0.0:7777

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000`
