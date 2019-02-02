package fly

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
)

// AWSPipeline is AWS specific implementation of Pipeline interface
type AWSPipeline struct {
	PipelineTemplateParams
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	credsGetter AWSCredsGetter
}

// NewAWSPipeline return AWSPipeline
func NewAWSPipeline(getter AWSCredsGetter) Pipeline {
	return AWSPipeline{credsGetter: getter}
}

type AWSCredsGetter = func()(string, string, error)
var getCredsFromSession = func() (string, string, error) {
	sess, err := session.NewSession()
	if err != nil {
		return "", "", err
	}

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return "", "", err
	}

	return creds.AccessKeyID, creds.SecretAccessKey, nil
}

//BuildPipelineParams builds params for AWS concourse-up self update pipeline
func (a AWSPipeline) BuildPipelineParams(deployment, namespace, region, domain string) (Pipeline, error) {
	accessKeyID, secretAccessKey, err := a.credsGetter()
	if err != nil {
		return nil, err
	}

	return AWSPipeline{
		PipelineTemplateParams: PipelineTemplateParams{
			ConcourseUpVersion: ConcourseUpVersion,
			Deployment:         strings.TrimPrefix(deployment, "concourse-up-"),
			Domain:             domain,
			Namespace:          namespace,
			Region:             region,
		},
		AWSAccessKeyID:     accessKeyID,
		AWSSecretAccessKey: secretAccessKey,
	}, nil
}

// GetConfigTemplate returns template for AWS Concourse Up self update pipeline
func (a AWSPipeline) GetConfigTemplate() string {
	return awsPipelineTemplate

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
      AWS_REGION: "{{ .Region }}"
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
      AWS_REGION: "{{ .Region }}"
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
