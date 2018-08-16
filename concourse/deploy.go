package concourse

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/EngineerBetter/concourse-up/iaas"

	"gopkg.in/yaml.v2"

	"strings"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/terraform"
)

// DeployAction runs Deploy
func (client *Client) DeployAction() error {
	_, err := client.Deploy()
	return err
}

// Deploy deploys a concourse instance
func (client *Client) Deploy() (config.Config, error) {
	c, err := client.loadConfig()
	if err != nil {
		return config.Config{}, err
	}
	isDomainUpdated := client.deployArgs.Domain != c.Domain

	c, err = client.checkPreTerraformConfigRequirements(c)
	if err != nil {
		return c, err
	}

	metadata, err := client.applyTerraform(c)
	if err != nil {
		return c, err
	}

	c, err = client.checkPreDeployConfigRequirements(client.acmeClientConstructor, isDomainUpdated, c, metadata)
	if err != nil {
		return c, err
	}

	if client.deployArgs.SelfUpdate {
		err = client.updateBoshAndPipeline(c, metadata)
	} else {
		err = client.deployBoshAndPipeline(c, metadata)
	}
	if err != nil {
		return c, err
	}
	return c, client.configClient.Update(c)
}

func (client *Client) deployBoshAndPipeline(config config.Config, metadata *terraform.Metadata) error {
	// When we are deploying for the first time rather than updating
	// ensure that the pipeline is set _after_ the concourse is deployed
	if err := client.deployBosh(config, metadata, false); err != nil {
		return err
	}

	flyClient, err := client.flyClientFactory(fly.Credentials{
		Target:   config.Deployment,
		API:      fmt.Sprintf("https://%s", config.Domain),
		Username: config.ConcourseUsername,
		Password: config.ConcoursePassword,
	},
		client.stdout,
		client.stderr,
		client.versionFile,
	)
	if err != nil {
		return err
	}
	defer flyClient.Cleanup()

	if err := flyClient.SetDefaultPipeline(client.deployArgs, config, false); err != nil {
		return err
	}

	return writeDeploySuccessMessage(config, metadata, client.stdout)
}

func (client *Client) updateBoshAndPipeline(c config.Config, metadata *terraform.Metadata) error {
	// If concourse is already running this is an update rather than a fresh deploy
	// When updating we need to deploy the BOSH as the final step in order to
	// Detach from the update, so the update job can exit
	flyClient, err := client.flyClientFactory(fly.Credentials{
		Target:   c.Deployment,
		API:      fmt.Sprintf("https://%s", c.Domain),
		Username: c.ConcourseUsername,
		Password: c.ConcoursePassword,
	},
		client.stdout,
		client.stderr,
		client.versionFile,
	)
	if err != nil {
		return err
	}
	defer flyClient.Cleanup()

	concourseAlreadyRunning, err := flyClient.CanConnect()
	if err != nil {
		return err
	}

	if !concourseAlreadyRunning {
		return fmt.Errorf("In detach mode but it seems that concourse is not currently running")
	}

	// Allow a fly version discrepancy since we might be targetting an older Concourse
	if err = flyClient.SetDefaultPipeline(client.deployArgs, c, true); err != nil {
		return err
	}

	if err = client.deployBosh(c, metadata, true); err != nil {
		return err
	}

	_, err = client.stdout.Write([]byte("\nUPGRADE RUNNING IN BACKGROUND\n\n"))

	return err
}

func (client *Client) checkPreTerraformConfigRequirements(conf config.Config) (config.Config, error) {
	region := client.deployArgs.AWSRegion

	if conf.Region != "" {
		if conf.Region != region {
			return conf, fmt.Errorf("found previous deployment in %s. Refusing to deploy to %s as changing regions for existing deployments is not supported", conf.Region, region)
		}
	}

	conf.Region = region

	// If the RDS instance size has manually set, override the existing size in the config
	if client.deployArgs.DBSizeIsSet {
		conf.RDSInstanceClass = config.DBSizes[client.deployArgs.DBSize]
	}

	// When in self-update mode do not override the user IP, since we already have access to the worker
	if !client.deployArgs.SelfUpdate {
		var err error
		conf, err = client.setUserIP(conf)
		if err != nil {
			return conf, err
		}
	}

	conf, err := client.setHostedZone(conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func (client *Client) checkPreDeployConfigRequirements(c func(u *certs.User) (certs.AcmeClient, error), isDomainUpdated bool, cfg config.Config, metadata *terraform.Metadata) (config.Config, error) {
	if client.deployArgs.Domain == "" {
		cfg.Domain = metadata.ATCPublicIP.Value
	}

	cfg, err := client.ensureDirectorCerts(c, cfg, metadata)
	if err != nil {
		return config.Config{}, err
	}

	cfg, err = client.ensureConcourseCerts(c, isDomainUpdated, cfg, metadata)
	if err != nil {
		return config.Config{}, err
	}

	cfg.ConcourseWorkerCount = client.deployArgs.WorkerCount
	cfg.ConcourseWorkerSize = client.deployArgs.WorkerSize
	cfg.ConcourseWebSize = client.deployArgs.WebSize
	cfg.DirectorPublicIP = metadata.DirectorPublicIP.Value

	if err := client.configClient.Update(cfg); err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

func (client *Client) ensureDirectorCerts(c func(u *certs.User) (certs.AcmeClient, error), cfg config.Config, metadata *terraform.Metadata) (config.Config, error) {
	// If we already have director certificates, don't regenerate as changing them will
	// force a bosh director re-deploy even if there are no other changes
	if cfg.DirectorCACert != "" {
		return cfg, nil
	}

	ip := metadata.DirectorPublicIP.Value
	_, err := client.stdout.Write(
		[]byte(fmt.Sprintf("\nGENERATING BOSH DIRECTOR CERTIFICATE (%s, 10.0.0.6)\n", ip)))
	if err != nil {
		return config.Config{}, err
	}

	directorCerts, err := client.certGenerator(c, cfg.Deployment, ip, "10.0.0.6")
	if err != nil {
		return config.Config{}, err
	}

	cfg.DirectorCACert = string(directorCerts.CACert)
	cfg.DirectorCert = string(directorCerts.Cert)
	cfg.DirectorKey = string(directorCerts.Key)

	return cfg, nil
}

func timeTillExpiry(cert string) time.Duration {
	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		return 0
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return 0
	}
	return time.Until(c.NotAfter)
}

func (client *Client) ensureConcourseCerts(c func(u *certs.User) (certs.AcmeClient, error), domainUpdated bool, cfg config.Config, metadata *terraform.Metadata) (config.Config, error) {
	if client.deployArgs.TLSCert != "" {
		cfg.ConcourseCert = client.deployArgs.TLSCert
		cfg.ConcourseKey = client.deployArgs.TLSKey
		cfg.ConcourseUserProvidedCert = true

		return cfg, nil
	}

	// Skip concourse re-deploy if certs have already been set,
	// unless domain has changed
	if cfg.ConcourseCert != "" && !domainUpdated && timeTillExpiry(cfg.ConcourseCert) > 28*24*time.Hour {
		return cfg, nil
	}

	// If no domain has been provided by the user, the value of cfg.Domain is set to the ATC's public IP in checkPreDeployConfigRequirements
	concourseCerts, err := client.certGenerator(c, cfg.Deployment, cfg.Domain)
	if err != nil {
		return config.Config{}, err
	}

	cfg.ConcourseCert = string(concourseCerts.Cert)
	cfg.ConcourseKey = string(concourseCerts.Key)
	cfg.ConcourseCACert = string(concourseCerts.CACert)

	return cfg, nil
}

func (client *Client) applyTerraform(c config.Config) (*terraform.Metadata, error) {
	terraformClient, err := client.terraformClientFactory(client.iaasClient.IAAS(), c, client.stdout, client.stderr, client.versionFile)
	if err != nil {
		return nil, err
	}
	defer terraformClient.Cleanup()

	if err = terraformClient.Apply(false); err != nil {
		return nil, err
	}

	metadata, err := terraformClient.Output()
	if err != nil {
		return nil, err
	}

	if err = metadata.AssertValid(); err != nil {
		return nil, err
	}

	return metadata, nil
}

func (client *Client) deployBosh(config config.Config, metadata *terraform.Metadata, detach bool) error {
	boshClient, err := client.buildBoshClient(config, metadata)
	if err != nil {
		return err
	}
	defer boshClient.Cleanup()

	boshStateBytes, err := loadDirectorState(client.configClient)
	if err != nil {
		return err
	}
	boshCredsBytes, err := loadDirectorCreds(client.configClient)
	if err != nil {
		return err
	}

	boshStateBytes, boshCredsBytes, err = boshClient.Deploy(boshStateBytes, boshCredsBytes, detach)
	err1 := client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
	if err == nil {
		err = err1
	}
	err1 = client.configClient.StoreAsset(bosh.CredsFilename, boshCredsBytes)
	if err == nil {
		err = err1
	}
	if err != nil {
		return err
	}

	var cc struct {
		CredhubPassword          string `yaml:"credhub_cli_password"`
		CredhubAdminClientSecret string `yaml:"credhub_admin_client_secret"`
		InternalTLS              struct {
			CA string `yaml:"ca"`
		} `yaml:"internal_tls"`
		AtcPassword string `yaml:"atc_password"`
	}

	err = yaml.Unmarshal(boshCredsBytes, &cc)
	if err != nil {
		return err
	}

	config.CredhubPassword = cc.CredhubPassword
	config.CredhubAdminClientSecret = cc.CredhubAdminClientSecret
	config.CredhubCACert = cc.InternalTLS.CA
	config.CredhubURL = fmt.Sprintf("https://%s:8844/", config.Domain)
	config.CredhubUsername = "credhub-cli"
	config.ConcourseUsername = "admin"
	if len(cc.AtcPassword) > 0 {
		config.ConcoursePassword = cc.AtcPassword
		config.GrafanaPassword = cc.AtcPassword
	}

	return nil
}

func (client *Client) loadConfig() (config.Config, error) {
	cfg, createdNewConfig, err := client.configClient.LoadOrCreate(client.deployArgs)
	if err != nil {
		return config.Config{}, err
	}

	if !createdNewConfig {
		if err = writeConfigLoadedSuccessMessage(client.stdout); err != nil {
			return config.Config{}, err
		}
	}
	return cfg, nil
}

func (client *Client) setUserIP(c config.Config) (config.Config, error) {
	userIP, err := client.ipChecker()
	if err != nil {
		return c, err
	}

	if c.SourceAccessIP != userIP {
		c.SourceAccessIP = userIP
		_, err = client.stderr.Write([]byte(fmt.Sprintf(
			"\nWARNING: allowing access from local machine (address: %s)\n\n", userIP)))
		if err != nil {
			return c, err
		}
		if err = client.configClient.Update(c); err != nil {
			return c, err
		}
	}

	return c, nil
}

func (client *Client) setHostedZone(c config.Config) (config.Config, error) {
	domain := client.deployArgs.Domain
	if client.deployArgs.Domain == "" {
		return c, nil
	}

	hostedZoneName, hostedZoneID, err := client.iaasClient.FindLongestMatchingHostedZone(domain, iaas.ListHostedZones)
	if err != nil {
		return c, err
	}
	c.HostedZoneID = hostedZoneID
	c.HostedZoneRecordPrefix = strings.TrimSuffix(domain, fmt.Sprintf(".%s", hostedZoneName))
	c.Domain = domain

	_, err = client.stderr.Write([]byte(fmt.Sprintf(
		"\nWARNING: adding record %s to Route53 hosted zone %s ID: %s\n\n", domain, hostedZoneName, hostedZoneID)))
	if err != nil {
		return c, err
	}
	err = client.configClient.Update(c)

	return c, err
}

const deployMsg = `DEPLOY SUCCESSFUL. Log in with:
fly --target {{.Project}} login{{if not .ConcourseUserProvidedCert}} --insecure{{end}} --concourse-url https://{{.Domain}} --username {{.ConcourseUsername}} --password {{.ConcoursePassword}}

Metrics available at https://{{.Domain}}:3000 using the same username and password

Log into credhub with:
eval "$(concourse-up info {{.Project}} --region {{.Region}} --env)"
`

func writeDeploySuccessMessage(config config.Config, metadata *terraform.Metadata, stdout io.Writer) error {
	t := template.Must(template.New("deploy").Parse(deployMsg))
	return t.Execute(stdout, config)
}

func writeConfigLoadedSuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nUSING PREVIOUS DEPLOYMENT CONFIG\n"))

	return err
}

func loadDirectorState(configClient config.IClient) ([]byte, error) {
	hasState, err := configClient.HasAsset(bosh.StateFilename)
	if err != nil {
		return nil, err
	}

	if !hasState {
		return nil, nil
	}

	return configClient.LoadAsset(bosh.StateFilename)
}
func loadDirectorCreds(configClient config.IClient) ([]byte, error) {
	hasCreds, err := configClient.HasAsset(bosh.CredsFilename)
	if err != nil {
		return nil, err
	}

	if !hasCreds {
		return nil, nil
	}

	return configClient.LoadAsset(bosh.CredsFilename)
}
