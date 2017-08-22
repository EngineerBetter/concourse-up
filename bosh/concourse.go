package bosh

import (
	"fmt"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/db"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

const concourseManifestFilename = "concourse.yml"
const concourseDeploymentName = "concourse"

var ConcourseStemcellURL = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellURL"
var ConcourseStemcellVersion = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion"
var ConcourseStemcellSHA1 = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellSHA1"
var ConcourseReleaseURL = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseURL"
var ConcourseReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseVersion"
var ConcourseReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseSHA1"
var GardenReleaseURL = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseURL"
var GardenReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseVersion"
var GardenReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseSHA1"

func (client *Client) uploadConcourseStemcell() error {
	if err := client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		"upload-stemcell",
		ConcourseStemcellURL,
	); err != nil {
		return err
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

	if err = client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		"--deployment",
		concourseDeploymentName,
		"deploy",
		concourseManifestPath,
	); err != nil {
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
		ConcourseReleaseVersion: ConcourseReleaseVersion,
		ConcourseReleaseSHA1:    ConcourseReleaseSHA1,
		ConcourseReleaseURL:     ConcourseReleaseURL,
		GardenReleaseVersion:    GardenReleaseVersion,
		GardenReleaseURL:        GardenReleaseURL,
		GardenReleaseSHA1:       GardenReleaseSHA1,
		StemcellVersion:         ConcourseStemcellVersion,
		StemcellSHA1:            ConcourseStemcellSHA1,
		StemcellURL:             ConcourseStemcellURL,
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
	ConcourseReleaseSHA1    string
	ConcourseReleaseURL     string
	GardenReleaseVersion    string
	GardenReleaseURL        string
	GardenReleaseSHA1       string
	StemcellVersion         string
	StemcellSHA1            string
	StemcellURL             string
}

// Indent is a helper function to indent the field a given number of spaces
func (params awsConcourseManifestParams) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

const awsConcourseManifestTemplate = `---
name: concourse

releases:
- name: concourse
  url: "<% .ConcourseReleaseURL %>"
  sha1: "<% .ConcourseReleaseSHA1 %>"
  version: <% .ConcourseReleaseVersion %>

- name: garden-runc
  url: "<% .GardenReleaseURL %>"
  sha1: "<% .GardenReleaseSHA1 %>"
  version: <% .GardenReleaseVersion %>

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: "<% .StemcellVersion %>"

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
