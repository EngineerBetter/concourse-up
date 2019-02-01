package fly

import (
	"strings"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/aws/aws-sdk-go/aws/session"
)

// AWSPipeline is AWS specific implementation of Pipeline interface
type AWSPipeline struct {
	AWSAccessKeyID       string
	AWSDefaultRegion     string
	AWSSecretAccessKey   string
	Deployment           string
	FlagAWSRegion        string
	FlagDomain           string
	FlagGithubAuthID     string
	FlagGithubAuthSecret string
	FlagTLSCert          string
	FlagTLSKey           string
	FlagWebSize          string
	FlagWorkerSize       string
	FlagWorkers          int
	ConcourseUpVersion   string
	Namespace            string
}

// NewAWSPipeline return AWSPipeline
func NewAWSPipeline() Pipeline {
	return AWSPipeline{}
}

//BuildPipelineParams builds params for AWS concourse-up self update pipeline
func (a AWSPipeline) BuildPipelineParams(config config.Config) (Pipeline, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	var (
		domain        string
		concourseCert string
		concourseKey  string
	)

	if !validIP4(config.Domain) {
		domain = config.Domain
	}

	if domain != "" {
		concourseCert = config.ConcourseCert
		concourseKey = config.ConcourseKey
	}

	return AWSPipeline{
		AWSAccessKeyID:       creds.AccessKeyID,
		AWSSecretAccessKey:   creds.SecretAccessKey,
		Deployment:           strings.TrimPrefix(config.Deployment, "concourse-up-"),
		FlagAWSRegion:        config.Region,
		FlagDomain:           domain,
		FlagGithubAuthID:     config.GithubClientID,
		FlagGithubAuthSecret: config.GithubClientSecret,
		FlagTLSCert:          concourseCert,
		FlagTLSKey:           concourseKey,
		FlagWebSize:          config.ConcourseWebSize,
		FlagWorkerSize:       config.ConcourseWorkerSize,
		FlagWorkers:          config.ConcourseWorkerCount,
		ConcourseUpVersion:   ConcourseUpVersion,
		Namespace:            config.Namespace,
	}, nil
}

// GetConfigTemplate returns template for AWS Concourse Up self update pipeline
func (a AWSPipeline) GetConfigTemplate() string {
	return awsPipelineTemplate

}

const awsPipelineTemplate = `
---
` + selfUpdateResources + `
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
      DOMAIN: "{{ .FlagDomain }}"
      TLS_CERT: |-
        {{ .Indent "8" .FlagTLSCert }}
      TLS_KEY: |-
        {{ .Indent "8" .FlagTLSKey }}
      WORKERS: "{{ .FlagWorkers }}"
      WORKER_SIZE: "{{ .FlagWorkerSize }}"
      WEB_SIZE: "{{ .FlagWebSize }}"
      DEPLOYMENT: "{{ .Deployment }}"
      GITHUB_AUTH_CLIENT_ID: "{{ .FlagGithubAuthID }}"
      GITHUB_AUTH_CLIENT_SECRET: "{{ .FlagGithubAuthSecret }}"
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
- name: renew-cert
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
      DOMAIN: "{{ .FlagDomain }}"
      TLS_CERT: |-
        {{ .Indent "8" .FlagTLSCert }}
      TLS_KEY: |-
        {{ .Indent "8" .FlagTLSKey }}
      WORKERS: "{{ .FlagWorkers }}"
      WORKER_SIZE: "{{ .FlagWorkerSize }}"
      WEB_SIZE: "{{ .FlagWebSize }}"
      DEPLOYMENT: "{{ .Deployment }}"
      GITHUB_AUTH_CLIENT_ID: "{{ .FlagGithubAuthID }}"
      GITHUB_AUTH_CLIENT_SECRET: "{{ .FlagGithubAuthSecret }}"
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
