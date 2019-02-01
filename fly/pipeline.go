package fly

// Pipeline is interface for self update pipeline
type Pipeline interface {
	BuildPipelineParams(deployment, namespace, region, domain string) (Pipeline, error)
	GetConfigTemplate() string
}

type PipelineTemplateParams struct {
	ConcourseUpVersion string
	Deployment         string
	Domain             string
	Namespace          string
	Region             string
}

const selfUpdateResources = `
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
`

const renewCertsDateCheck = `
          now_seconds=$(date +%s)
          not_after=$(echo | openssl s_client -connect {{.Domain}}:443 2>/dev/null | openssl x509 -noout -enddate)
          expires_on=${not_after#'notAfter='}
          expires_on_seconds=$(date --date="$expires_on" +%s)
          let "seconds_until_expiry = $expires_on_seconds - $now_seconds"
          let "days_until_expiry = $seconds_until_expiry / 60 / 60 / 24"
          if [ $days_until_expiry -gt 2 ]; then
            echo Not renewing HTTPS cert, as they do not expire in the next two days.
            exit 0
          fi
`
