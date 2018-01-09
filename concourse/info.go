package concourse

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/fatih/color"
)

// Info represents the terraform output and concourse-up config files
type Info struct {
	Terraform *terraform.Metadata `json:"terraform"`
	Config    *config.Config      `json:"config"`
	Instances []bosh.Instance     `json:"instances"`
}

// FetchInfo fetches and builds the info
func (client *Client) FetchInfo() (*Info, error) {
	config, err := client.configClient.Load()
	if err != nil {
		return nil, err
	}

	terraformClient, err := client.terraformClientFactory(client.iaasClient.IAAS(), config, client.stdout, client.stderr)
	if err != nil {
		return nil, err
	}
	defer terraformClient.Cleanup()

	metadata, err := terraformClient.Output()
	if err != nil {
		return nil, err
	}

	boshClient, err := client.buildBoshClient(config, metadata)
	if err != nil {
		return nil, err
	}
	defer boshClient.Cleanup()

	instances, err := boshClient.Instances()
	if err != nil {
		return nil, err
	}

	return &Info{
		Terraform: metadata,
		Config:    config,
		Instances: instances,
	}, nil
}

const infoTemplate = `Deployment:
	IAAS:   aws
	Region: {{.Config.Region}}

Workers:
	Count:              {{.Config.ConcourseWorkerCount}}
	Size:               {{.Config.ConcourseWorkerSize}}
	Outbound Public IP: {{.Terraform.NatGatewayIP.Value}}

Instances:
{{range .Instances}}
	{{.Name}} {{.IP | replace "\n" ","}} {{.State}}
{{end}}

Concourse credentials:
	username: {{.Config.ConcourseUsername}}
	password: {{.Config.ConcoursePassword}}
	URL:      https://{{.Config.Domain}}

Credhub credentials:
	username: {{.Config.CredhubUsername}}
	password: {{.Config.CredhubPassword}}
	URL:      {{.Config.CredhubURL}}
	CA Cert:
		{{ .Config.CredhubCACert | replace "\n" "\n\t\t"}}
		
Grafana credentials:
	username: {{.Config.ConcourseUsername}}
	password: {{.Config.ConcoursePassword}}
	URL:      https://{{.Config.Domain}}:3000

Bosh credentials:
	username: {{.Config.DirectorUsername}}
	password: {{.Config.DirectorPassword}}
	IP:       {{.Terraform.DirectorPublicIP.Value}}
	CA Cert:
		{{ .Config.DirectorCACert | replace "\n" "\n\t\t"}}
		
Built by {{"EngineerBetter http://engineerbetter.com" | blue}}`

func (info *Info) String() string {
	t := template.Must(template.New("info").Funcs(template.FuncMap{
		"replace": func(old, new, s string) string {
			return strings.Replace(s, old, new, -1)
		},
		"blue": color.New(color.FgCyan, color.Bold).Sprint,
	}).Parse(infoTemplate))
	var buf bytes.Buffer
	err := t.Execute(&buf, info)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
