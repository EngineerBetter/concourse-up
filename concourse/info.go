package concourse

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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

	terraformClient, err := client.terraformClientFactory(client.iaasClient.IAAS(), config, client.stdout, client.stderr, client.versionFile)
	if err != nil {
		return nil, err
	}
	defer terraformClient.Cleanup()

	metadata, err := terraformClient.Output()
	if err != nil {
		return nil, err
	}

	userIP, err := client.ipChecker()
	if err != nil {
		return nil, err
	}

	whitelisted, err := client.iaasClient.CheckForWhitelistedIP(userIP, metadata.DirectorSecurityGroupID.Value, client.iaasClient.NewEC2Client)
	if err != nil {
		return nil, err
	}

	if !whitelisted {
		err = fmt.Errorf("Do you need to add your IP %s to the %s-director security group (for ports 22, 6868, and 25555)?", userIP, config.Deployment)
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

Built by {{"EngineerBetter http://engineerbetter.com" | blue}}
`

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

func writeTempFile(data string) (name string, err error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	name = f.Name()
	_, err = f.WriteString(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(name)
	}
	return name, err
}

var envTemplate = template.Must(template.New("env").Funcs(template.FuncMap{
	"to_file": writeTempFile,
}).Parse(`
export BOSH_CA_CERT='{{.Config.DirectorCACert}}'
export BOSH_ENVIRONMENT={{.Terraform.DirectorPublicIP.Value}}
export BOSH_DEPLOYMENT=concourse
export BOSH_CLIENT={{.Config.DirectorUsername}}
export BOSH_CLIENT_SECRET={{.Config.DirectorPassword}}
export BOSH_GW_USER=vcap
export BOSH_GW_HOST={{.Terraform.DirectorPublicIP.Value}}
export BOSH_GW_PRIVATE_KEY={{.Config.PrivateKey | to_file}}
export CREDHUB_SERVER={{.Config.CredhubURL}}
export CREDHUB_CA_CERT='{{.Config.CredhubCACert}}'
export CREDHUB_CLIENT=credhub_admin
export CREDHUB_SECRET={{.Config.CredhubAdminClientSecret}}
`))

// Env returns a string that is suitable for a shell to evaluate that sets environment
// varibles which are used to log into bosh and credhub
func (info *Info) Env() (string, error) {
	var buf bytes.Buffer
	err := envTemplate.Execute(&buf, info)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
