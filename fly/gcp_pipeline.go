package fly

import (
	"io/ioutil"
	"strings"

	"github.com/EngineerBetter/concourse-up/util"
)

// GCPPipeline is GCP specific implementation of Pipeline interface
type GCPPipeline struct {
	PipelineTemplateParams
	GCPCreds string
}

// NewGCPPipeline return GCPPipeline
func NewGCPPipeline(credsPath string) (Pipeline, error) {
	creds, err := readFileContents(credsPath)
	if err != nil {
		return nil, err
	}
	return GCPPipeline{
		GCPCreds: creds,
	}, nil
}

//BuildPipelineParams builds params for AWS concourse-up self update pipeline
func (a GCPPipeline) BuildPipelineParams(deployment, namespace, region, domain string) (Pipeline, error) {
	return GCPPipeline{
		PipelineTemplateParams: PipelineTemplateParams{
			ConcourseUpVersion: ConcourseUpVersion,
			Deployment:         strings.TrimPrefix(deployment, "concourse-up-"),
			Domain:             domain,
			Namespace:          namespace,
			Region:             region,
		},
		GCPCreds: a.GCPCreds,
	}, nil
}

// GetConfigTemplate returns template for AWS Concourse Up self update pipeline
func (a GCPPipeline) GetConfigTemplate() string {
	return gcpPipelineTemplate

}

// Indent is a helper function to indent the field a given number of spaces
func (a GCPPipeline) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

func readFileContents(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

const gcpPipelineTemplate = `
---` + selfUpdateResources + `
jobs:
- name: self-update
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    trigger: true
  - task: update
    params:
      AWS_REGION: "{{ .Region }}"
      DEPLOYMENT: "{{ .Deployment }}"
      IAAS: GCP
      SELF_UPDATE: true
      NAMESPACE: {{ .Namespace }}
      GCPCreds: '{{ .GCPCreds }}'
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/pcf-ops
      inputs:
      - name: concourse-up-release
      run:
        path: bash
        args:
        - -c
        - |
          cd concourse-up-release
          echo "${GCPCreds}" > googlecreds.json
          export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
          set -eux
          chmod +x concourse-up-linux-amd64
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
- name: renew-https-cert
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    version: {tag: "{{ .ConcourseUpVersion }}" }
  - get: every-day
    trigger: true
  - task: update
    params:
      AWS_REGION: "{{ .Region }}"
      DEPLOYMENT: "{{ .Deployment }}"
      IAAS: GCP
      SELF_UPDATE: true
      NAMESPACE: "{{ .Namespace }}"
      GCPCreds: '{{ .GCPCreds }}'
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/pcf-ops
      inputs:
      - name: concourse-up-release
      run:
        path: bash
        args:
        - -c
        - |
          echo "${GCPCreds}" > googlecreds.json
          export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
          set -euxo pipefail
          cd concourse-up-release
          chmod +x concourse-up-linux-amd64
` + renewCertsDateCheck + `
          echo Certificates expire in $days_until_expiry days, redeploying to renew them
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
`
