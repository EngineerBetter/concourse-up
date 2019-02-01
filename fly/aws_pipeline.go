package fly

import (
	"strings"

	"github.com/EngineerBetter/concourse-up/util"
	"github.com/aws/aws-sdk-go/aws/session"
)

// AWSPipeline is AWS specific implementation of Pipeline interface
type AWSPipeline struct {
	AWSAccessKeyID     string
	AWSDefaultRegion   string
	AWSSecretAccessKey string
	Deployment         string
	Domain             string
	FlagAWSRegion      string
	ConcourseUpVersion string
	Namespace          string
}

// NewAWSPipeline return AWSPipeline
func NewAWSPipeline() Pipeline {
	return AWSPipeline{}
}

//BuildPipelineParams builds params for AWS concourse-up self update pipeline
func (a AWSPipeline) BuildPipelineParams(deployment, namespace, region, domain string) (Pipeline, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	return AWSPipeline{
		AWSAccessKeyID:     creds.AccessKeyID,
		AWSSecretAccessKey: creds.SecretAccessKey,
		Deployment:         strings.TrimPrefix(deployment, "concourse-up-"),
		Domain:             domain,
		FlagAWSRegion:      region,
		ConcourseUpVersion: ConcourseUpVersion,
		Namespace:          namespace,
	}, nil
}

// GetConfigTemplate returns template for AWS Concourse Up self update pipeline
func (a AWSPipeline) GetConfigTemplate() string {
	return awsPipelineTemplate

}

// Indent is a helper function to indent the field a given number of spaces
func (a AWSPipeline) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

const awsPipelineTemplate = `
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
      AWS_REGION: "{{ .FlagAWSRegion }}"
      DEPLOYMENT: "{{ .Deployment }}"
      AWS_ACCESS_KEY_ID: "{{ .AWSAccessKeyID }}"
      AWS_SECRET_ACCESS_KEY: "{{ .AWSSecretAccessKey }}"
      SELF_UPDATE: true
      NAMESPACE: {{ .Namespace }}
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
          set -eux

          cd concourse-up-release
          chmod +x concourse-up-linux-amd64
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
- name: renew-https-cert
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    version: {tag: {{ .ConcourseUpVersion }} }
  - get: every-day
    trigger: true
  - task: update
    params:
      AWS_REGION: "{{ .FlagAWSRegion }}"
      DEPLOYMENT: "{{ .Deployment }}"
      AWS_ACCESS_KEY_ID: "{{ .AWSAccessKeyID }}"
      AWS_SECRET_ACCESS_KEY: "{{ .AWSSecretAccessKey }}"
      SELF_UPDATE: true
      NAMESPACE: {{ .Namespace }}
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
          set -euxo pipefail
          cd concourse-up-release
          chmod +x concourse-up-linux-amd64
` + renewCertsDateCheck + `
          echo Certificates expire in $days_until_expiry days, redeploying to renew them
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
`
