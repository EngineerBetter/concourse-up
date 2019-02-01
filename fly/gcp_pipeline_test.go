package fly_test

import (
	. "github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

var _ = Describe("GCPPipeline", func() {
	Describe("Generating a pipeline YAML", func() {
		It("Generates something sensible", func() {
			tempFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			defer os.Remove(tempFile.Name()) // clean up

			_, err = tempFile.Write([]byte("creds-content"))
			Expect(err).ToNot(HaveOccurred())

			pipeline, err := NewGCPPipeline(tempFile.Name())
			Expect(err).ToNot(HaveOccurred())

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "europe-west1", "ci.engineerbetter.com")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate(pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			actual := string(yamlBytes)
			Expect(actual).To(Equal(expected))
		})
	})
})

var expected = `
---
resources:
- name: concourse-up-release
  type: github-release
  source:
    user: engineerbetter
    repository: concourse-up
    pre_release: true
- name: every-day
  type: time
  source: {interval: 24h}

jobs:
- name: self-update
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    trigger: true
  - task: update
    params:
      AWS_REGION: "europe-west1"
      DEPLOYMENT: "my-deployment"
      IAAS: GCP
      SELF_UPDATE: true
      NAMESPACE: prod
      GCPCreds: 'creds-content'
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
    version: {tag: "COMPILE_TIME_VARIABLE_fly_concourse_up_version" }
  - get: every-day
    trigger: true
  - task: update
    params:
      AWS_REGION: "europe-west1"
      DEPLOYMENT: "my-deployment"
      IAAS: GCP
      SELF_UPDATE: true
      NAMESPACE: "prod"
      GCPCreds: 'creds-content'
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

          now_seconds=$(date +%s)
          not_after=$(echo | openssl s_client -connect ci.engineerbetter.com:443 2>/dev/null | openssl x509 -noout -enddate)
          expires_on=${not_after#'notAfter='}
          expires_on_seconds=$(date --date="$expires_on" +%s)
          let "seconds_until_expiry = $expires_on_seconds - $now_seconds"
          let "days_until_expiry = $seconds_until_expiry / 60 / 60 / 24"
          if [ $days_until_expiry -gt 2 ]; then
            echo Not renewing HTTPS cert, as they do not expire in the next two days.
            exit 0
          fi

          echo Certificates expire in $days_until_expiry days, redeploying to renew them
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
`
