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
	"github.com/fatih/color"
)

// Info represents the compound fields for info templates
type Info struct {
	Terraform TerraformInfo   `json:"terraform"`
	Config    config.Config   `json:"config"`
	Instances []bosh.Instance `json:"instances"`
}

// TerraformInfo represents the terraform output fields needed for the info templates
type TerraformInfo struct {
	DirectorPublicIP string
	NatGatewayIP     string
}

// FetchInfo fetches and builds the info
func (client *Client) FetchInfo() (*Info, error) {
	config, err := client.configClient.Load()
	if err != nil {
		return nil, err
	}

	environment, metadata, err := client.tfCLI.IAAS(client.provider.IAAS())
	if err != nil {
		return nil, err
	}
	switch client.provider.IAAS() {
	case "AWS": // nolint
		config.RDSDefaultDatabaseName = fmt.Sprintf("bosh_%s", eightRandomLetters())

		err = environment.Build(map[string]interface{}{
			"AllowIPs":               config.AllowIPs,
			"AvailabilityZone":       config.AvailabilityZone,
			"ConfigBucket":           config.ConfigBucket,
			"Deployment":             config.Deployment,
			"HostedZoneID":           config.HostedZoneID,
			"HostedZoneRecordPrefix": config.HostedZoneRecordPrefix,
			"Namespace":              config.Namespace,
			"Project":                config.Project,
			"PublicKey":              config.PublicKey,
			"RDSDefaultDatabaseName": config.RDSDefaultDatabaseName,
			"RDSInstanceClass":       config.RDSInstanceClass,
			"RDSPassword":            config.RDSPassword,
			"RDSUsername":            config.RDSUsername,
			"Region":                 config.Region,
			"SourceAccessIP":         config.SourceAccessIP,
			"TFStatePath":            config.TFStatePath,
			"MultiAZRDS":             config.MultiAZRDS,
		})
		if err != nil {
			return nil, err
		}
	case "GCP": // nolint
		config.RDSDefaultDatabaseName = fmt.Sprintf("bosh-%s", eightRandomLetters())

		project, err1 := client.provider.Attr("project")
		if err1 != nil {
			return nil, err1
		}
		credentialspath, err1 := client.provider.Attr("credentials_path")
		if err1 != nil {
			return nil, err1
		}
		err1 = environment.Build(map[string]interface{}{
			"Region":             client.provider.Region(),
			"Zone":               client.provider.Zone(""),
			"Tags":               "",
			"Project":            project,
			"GCPCredentialsJSON": credentialspath,
			"ExternalIP":         config.SourceAccessIP,
			"Deployment":         config.Deployment,
			"ConfigBucket":       config.ConfigBucket,
			"DBName":             config.RDSDefaultDatabaseName,
			"DBTier":             "db-f1-micro",
			"DBPassword":         config.RDSPassword,
			"DBUsername":         config.RDSUsername,
		})
		if err1 != nil {
			return nil, err1
		}
	}
	err = client.tfCLI.BuildOutput(environment, metadata)
	if err != nil {
		return nil, err
	}

	directorPublicIP, err := metadata.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}

	natGatewayIP, err := metadata.Get("NatGatewayIP")
	if err != nil {
		return nil, err
	}

	terraformInfo := TerraformInfo{
		DirectorPublicIP: directorPublicIP,
		NatGatewayIP:     natGatewayIP,
	}

	userIP, err := client.ipChecker()
	if err != nil {
		return nil, err
	}

	directorSecurityGroupID, err := metadata.Get("DirectorSecurityGroupID")
	if err != nil {
		return nil, err
	}
	whitelisted, err := client.provider.CheckForWhitelistedIP(userIP, directorSecurityGroupID)
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
		Terraform: terraformInfo,
		Config:    config,
		Instances: instances,
	}, nil
}

const infoTemplate = `Deployment:
	Namespace: {{.Config.Namespace}}
	IAAS:      aws
	Region:    {{.Config.Region}}

Workers:
	Count:              {{.Config.ConcourseWorkerCount}}
	Size:               {{.Config.ConcourseWorkerSize}}
	Outbound Public IP: {{.Terraform.NatGatewayIP}}

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
	IP:       {{.Terraform.DirectorPublicIP}}
	CA Cert:
		{{ .Config.DirectorCACert | replace "\n" "\n\t\t"}}

Uses Concourse-Up version {{.Config.Version}}

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
export BOSH_ENVIRONMENT={{.Terraform.DirectorPublicIP}}
export BOSH_GW_HOST={{.Terraform.DirectorPublicIP}}
export BOSH_CA_CERT='{{.Config.DirectorCACert}}'
export BOSH_DEPLOYMENT=concourse
export BOSH_CLIENT={{.Config.DirectorUsername}}
export BOSH_CLIENT_SECRET={{.Config.DirectorPassword}}
export BOSH_GW_USER=vcap
export BOSH_GW_PRIVATE_KEY={{.Config.PrivateKey | to_file}}
export CREDHUB_SERVER={{.Config.CredhubURL}}
export CREDHUB_CA_CERT='{{.Config.CredhubCACert}}'
export CREDHUB_CLIENT=credhub_admin
export CREDHUB_SECRET={{.Config.CredhubAdminClientSecret}}
export NAMESPACE={{.Config.Namespace}}
`))

// Env returns a string that is suitable for a shell to evaluate that sets environment
// varibles which are used to log into bosh and credhub
func (info *Info) Env() (string, error) {
	var buf bytes.Buffer
	var i Info
	i = *info
	err := envTemplate.Execute(&buf, i)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
