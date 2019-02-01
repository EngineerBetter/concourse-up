package fly

import "github.com/EngineerBetter/concourse-up/config"

// Pipeline is interface for self update pipeline
type Pipeline interface {
	BuildPipelineParams(config config.Config) (Pipeline, error)
	GetConfigTemplate() string
	Indent(countStr, field string) string
}

const selfUpdateResources = `
resources:
- name: concourse-up-release
  type: github-release
  source:
    user: engineerbetter
    repository: concourse-up
    pre_release: true
- name: every-month
  type: time
  source: {interval: 24h}
`

const renewCertsDateCheck = `
          now_seconds=$(date +%s)
          expires_on=$(concourse-up-linux-amd64 info $DEPLOYMENT --cert-expiry)
          expires_on_seconds=$(date --date="$expires_on" +%s)
          let "seconds_until_expiry = $expires_on_seconds - $now_seconds"
          let "days_until_expiry = $seconds_until_expiry / 60 / 60 / 24"
          if [ $days_until_expiry -gt 31 ]; then
            echo Not renewing certs, as they do not expire in the next month.
            exit 0
          fi
`
