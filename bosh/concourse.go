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

// ConcourseStemcellURL is a compile-time variable set with -ldflags
var ConcourseStemcellURL = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellURL"

// ConcourseStemcellVersion is a compile-time variable set with -ldflags
var ConcourseStemcellVersion = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellVersion"

// ConcourseStemcellSHA1 is a compile-time variable set with -ldflags
var ConcourseStemcellSHA1 = "COMPILE_TIME_VARIABLE_bosh_concourseStemcellSHA1"

// ConcourseReleaseURL is a compile-time variable set with -ldflags
var ConcourseReleaseURL = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseURL"

// ConcourseReleaseVersion is a compile-time variable set with -ldflags
var ConcourseReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseVersion"

// ConcourseReleaseSHA1 is a compile-time variable set with -ldflags
var ConcourseReleaseSHA1 = "COMPILE_TIME_VARIABLE_bosh_concourseReleaseSHA1"

// GardenReleaseURL is a compile-time variable set with -ldflags
var GardenReleaseURL = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseURL"

// GardenReleaseVersion is a compile-time variable set with -ldflags
var GardenReleaseVersion = "COMPILE_TIME_VARIABLE_bosh_gardenReleaseVersion"

// GardenReleaseSHA1 is a compile-time variable set with -ldflags
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
		AllowSelfSignedCerts:    "true",
		ConcourseReleaseSHA1:    ConcourseReleaseSHA1,
		ConcourseReleaseURL:     ConcourseReleaseURL,
		ConcourseReleaseVersion: ConcourseReleaseVersion,
		DBCACert:                db.RDSRootCert,
		DBHost:                  metadata.BoshDBAddress.Value,
		DBName:                  config.ConcourseDBName,
		DBPassword:              config.RDSPassword,
		DBPort:                  metadata.BoshDBPort.Value,
		DBUsername:              config.RDSUsername,
		EncryptionKey:           config.EncryptionKey,
		GardenReleaseSHA1:       GardenReleaseSHA1,
		GardenReleaseURL:        GardenReleaseURL,
		GardenReleaseVersion:    GardenReleaseVersion,
		Password:                config.ConcoursePassword,
		Project:                 config.Project,
		StemcellSHA1:            ConcourseStemcellSHA1,
		StemcellURL:             ConcourseStemcellURL,
		StemcellVersion:         ConcourseStemcellVersion,
		TLSCert:                 config.ConcourseCert,
		TLSKey:                  config.ConcourseKey,
		URL:                     fmt.Sprintf("https://%s", config.Domain),
		Username:                config.ConcourseUsername,
		WorkerCount:             config.ConcourseWorkerCount,
		WorkerSize:              config.ConcourseWorkerSize,
	}

	return util.RenderTemplate(awsConcourseManifestTemplate, templateParams)
}

type awsConcourseManifestParams struct {
	AllowSelfSignedCerts    string
	ConcourseReleaseSHA1    string
	ConcourseReleaseURL     string
	ConcourseReleaseVersion string
	DBCACert                string
	DBHost                  string
	DBName                  string
	DBPassword              string
	DBPort                  string
	DBUsername              string
	EncryptionKey           string
	GardenReleaseSHA1       string
	GardenReleaseURL        string
	GardenReleaseVersion    string
	Password                string
	Project                 string
	StemcellSHA1            string
	StemcellURL             string
	StemcellVersion         string
	TLSCert                 string
	TLSKey                  string
	URL                     string
	Username                string
	WorkerCount             int
	WorkerSize              string
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
  - name: public
    default: [dns, gateway]
  vm_extensions:
  - elb
  jobs:
  - name: atc
    release: concourse
    properties:
      allow_self_signed_certificates: <% .AllowSelfSignedCerts %>
      external_url: <% .URL %>
      encryption_key: <% .EncryptionKey %>
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
  - name: private
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
