package bosh

import (
	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
)

const concourseStemcellURL = "https://bosh-jenkins-artifacts.s3.amazonaws.com/bosh-stemcell/aws/light-bosh-stemcell-3262.4.1-aws-xen-ubuntu-trusty-go_agent.tgz"

var concourseReleaseURLs = []string{
	"https://bosh.io/d/github.com/concourse/concourse?v=2.7.3",
	"https://bosh.io/d/github.com/cloudfoundry/garden-runc-release?v=1.4.0",
}

func (client *Client) uploadConcourse() error {
	_, err := client.runAuthenticatedBoshCommand(
		"upload-stemcell",
		concourseStemcellURL,
	)
	if err != nil {
		return err
	}

	for _, releaseURL := range concourseReleaseURLs {
		_, err := client.runAuthenticatedBoshCommand(
			"upload-release",
			releaseURL,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateConcourseManifest(config *config.Config, metadata *terraform.Metadata) ([]byte, error) {
	templateParams := awsConcourseManifestParams{
		Workers:    config.ConcourseWorkerCount,
		URL:        metadata.ELBName.Value,
		Username:   config.ConcourseUsername,
		Password:   config.ConcoursePassword,
		DBUsername: config.RDSUsername,
		DBPassword: config.RDSPassword,
		DBName:     config.ConcourseDBName,
		DBHost:     metadata.BoshDBAddress.Value,
		DBCACert:   rdsRootCert,
		Network:    "default",
	}

	return util.RenderTemplate(awsConcourseManifestTemplate, templateParams)
}

type awsConcourseManifestParams struct {
	Workers    int
	URL        string
	Username   string
	Password   string
	DBHost     string
	DBName     string
	DBUsername string
	DBPassword string
	DBCACert   string
	Network    string
}

// Indent is a helper function to indent the field a given number of spaces
func (params awsConcourseManifestParams) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

const awsConcourseManifestTemplate = `---
name: concourse

releases:
- name: concourse
  version: latest
- name: garden-runc
  version: latest

stemcells:
- alias: trusty
  os: ubuntu-trusty
  version: latest

tags:
  project: concourse-up

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
      external_url: <% .URL %>
      basic_auth_username: <% .Username %>
      basic_auth_password: <% .Password %>
      postgresql:
        port: 5432
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
  instances: <% .Workers %>
  vm_type: concourse-large
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
