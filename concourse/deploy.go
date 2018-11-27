package concourse

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"text/template"
	"time"

	"strings"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/certs"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/terraform"
	"gopkg.in/yaml.v2"
)

// BoshParams represents the params used and produced by a BOSH deploy
type BoshParams struct {
	CredhubPassword          string
	CredhubAdminClientSecret string
	CredhubCACert            string
	CredhubURL               string
	CredhubUsername          string
	ConcourseUsername        string
	ConcoursePassword        string
	GrafanaPassword          string
	DirectorUsername         string
	DirectorPassword         string
	DirectorCACert           string
}

func stripVersion(tags []string) []string {
	output := []string{}
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "concourse-up-version") {
			output = append(output, tag)
		}
	}
	return output
}

// Deploy deploys a concourse instance
func (client *Client) Deploy() error {
	c, createdNewConfig, isDomainUpdated, err := client.configClient.LoadOrCreate(client.deployArgs)
	if err != nil {
		return err
	}

	if !createdNewConfig {
		if err = writeConfigLoadedSuccessMessage(client.stdout); err != nil {
			return err
		}
	} else {
		client.provider.WorkerType(c.ConcourseWorkerSize)
		c.AvailabilityZone = client.provider.Zone(client.deployArgs.Zone)
	}

	r, err := client.checkPreTerraformConfigRequirements(c, client.deployArgs.SelfUpdate)
	if err != nil {
		return err
	}
	c.Region = r.Region
	c.SourceAccessIP = r.SourceAccessIP
	c.HostedZoneID = r.HostedZoneID
	c.HostedZoneRecordPrefix = r.HostedZoneRecordPrefix
	c.Domain = r.Domain

	environment, metadata, err := client.tfCLI.IAAS(client.provider.IAAS())
	if err != nil {
		return err
	}
	switch client.provider.IAAS() {
	case "AWS": // nolint
		c.RDSDefaultDatabaseName = fmt.Sprintf("bosh_%s", eightRandomLetters())

		err = environment.Build(map[string]interface{}{
			"AllowIPs":               c.AllowIPs,
			"AvailabilityZone":       c.AvailabilityZone,
			"ConfigBucket":           c.ConfigBucket,
			"Deployment":             c.Deployment,
			"HostedZoneID":           c.HostedZoneID,
			"HostedZoneRecordPrefix": c.HostedZoneRecordPrefix,
			"Namespace":              c.Namespace,
			"Project":                c.Project,
			"PublicKey":              c.PublicKey,
			"RDSDefaultDatabaseName": c.RDSDefaultDatabaseName,
			"RDSInstanceClass":       c.RDSInstanceClass,
			"RDSPassword":            c.RDSPassword,
			"RDSUsername":            c.RDSUsername,
			"Region":                 c.Region,
			"SourceAccessIP":         c.SourceAccessIP,
			"TFStatePath":            c.TFStatePath,
			"MultiAZRDS":             c.MultiAZRDS,
		})
		if err != nil {
			return err
		}
	case "GCP": // nolint
		c.RDSDefaultDatabaseName = fmt.Sprintf("bosh-%s", eightRandomLetters())
		project, err1 := client.provider.Attr("project")
		if err1 != nil {
			return err1
		}
		credentialspath, err1 := client.provider.Attr("credentials_path")
		if err1 != nil {
			return err1
		}
		err1 = environment.Build(map[string]interface{}{
			"Region":             client.provider.Region(),
			"Zone":               client.provider.Zone(""),
			"Tags":               "",
			"Project":            project,
			"GCPCredentialsJSON": credentialspath,
			"ExternalIP":         c.SourceAccessIP,
			"Deployment":         c.Deployment,
			"ConfigBucket":       c.ConfigBucket,
			"DBTier":             "db-f1-micro",
			"DBPassword":         c.RDSPassword,
			"DBUsername":         c.RDSUsername,
			"DBName":             c.RDSDefaultDatabaseName,
		})
		if err1 != nil {
			return err1
		}
	default:
		return errors.New("concourse:deploy:unsupported iaas " + client.deployArgs.IAAS)
	}

	err = client.tfCLI.Apply(environment, false)
	if err != nil {
		return err
	}
	err = client.tfCLI.BuildOutput(environment, metadata)
	if err != nil {
		return err
	}

	err = client.configClient.Update(c)
	if err != nil {
		return err
	}
	c.Tags = stripVersion(c.Tags)
	c.Tags = append([]string{fmt.Sprintf("concourse-up-version=%s", client.version)}, c.Tags...)

	c.Version = client.version

	cr, err := client.checkPreDeployConfigRequirements(client.acmeClientConstructor, isDomainUpdated, c, metadata)
	if err != nil {
		return err
	}

	c.Domain = cr.Domain
	c.DirectorPublicIP = cr.DirectorPublicIP
	c.DirectorCACert = cr.DirectorCerts.DirectorCACert
	c.DirectorCert = cr.DirectorCerts.DirectorCert
	c.DirectorKey = cr.DirectorCerts.DirectorKey
	c.ConcourseCert = cr.Certs.ConcourseCert
	c.ConcourseKey = cr.Certs.ConcourseKey
	c.ConcourseUserProvidedCert = cr.Certs.ConcourseUserProvidedCert
	c.ConcourseCACert = cr.Certs.ConcourseCACert

	var bp BoshParams
	if client.deployArgs.SelfUpdate {
		bp, err = client.updateBoshAndPipeline(c, metadata)
	} else {
		bp, err = client.deployBoshAndPipeline(c, metadata)
	}
	if err != nil {
		return err
	}

	c.CredhubPassword = bp.CredhubPassword
	c.CredhubAdminClientSecret = bp.CredhubAdminClientSecret
	c.CredhubCACert = bp.CredhubCACert
	c.CredhubURL = bp.CredhubURL
	c.CredhubUsername = bp.CredhubUsername
	c.ConcourseUsername = bp.ConcourseUsername
	c.ConcoursePassword = bp.ConcoursePassword
	c.GrafanaPassword = bp.GrafanaPassword
	c.DirectorUsername = bp.DirectorUsername
	c.DirectorPassword = bp.DirectorPassword
	c.DirectorCACert = bp.DirectorCACert

	return client.configClient.Update(c)
}

func (client *Client) deployBoshAndPipeline(c config.Config, metadata terraform.IAASMetadata) (BoshParams, error) {
	// When we are deploying for the first time rather than updating
	// ensure that the pipeline is set _after_ the concourse is deployed

	bp := BoshParams{
		CredhubPassword:          c.CredhubPassword,
		CredhubAdminClientSecret: c.CredhubAdminClientSecret,
		CredhubCACert:            c.CredhubCACert,
		CredhubURL:               c.CredhubURL,
		CredhubUsername:          c.CredhubUsername,
		ConcourseUsername:        c.ConcourseUsername,
		ConcoursePassword:        c.ConcoursePassword,
		GrafanaPassword:          c.GrafanaPassword,
		DirectorUsername:         c.DirectorUsername,
		DirectorPassword:         c.DirectorPassword,
		DirectorCACert:           c.DirectorCACert,
	}

	bp, err := client.deployBosh(c, metadata, false)
	if err != nil {
		return bp, err
	}

	flyClient, err := client.flyClientFactory(fly.Credentials{
		Target:   c.Deployment,
		API:      fmt.Sprintf("https://%s", c.Domain),
		Username: bp.ConcourseUsername,
		Password: bp.ConcoursePassword,
	},
		client.stdout,
		client.stderr,
		client.versionFile,
	)
	if err != nil {
		return bp, err
	}
	defer flyClient.Cleanup()

	if err := flyClient.SetDefaultPipeline(c, false); err != nil {
		return bp, err
	}

	// This assignment is necessary for the deploy success message
	// It should be removed once we stop passing config everywhere
	c.ConcourseUsername = bp.ConcourseUsername
	c.ConcoursePassword = bp.ConcoursePassword

	return bp, writeDeploySuccessMessage(c, metadata, client.stdout)
}

func (client *Client) updateBoshAndPipeline(c config.Config, metadata terraform.IAASMetadata) (BoshParams, error) {
	// If concourse is already running this is an update rather than a fresh deploy
	// When updating we need to deploy the BOSH as the final step in order to
	// Detach from the update, so the update job can exit

	bp := BoshParams{
		CredhubPassword:          c.CredhubPassword,
		CredhubAdminClientSecret: c.CredhubAdminClientSecret,
		CredhubCACert:            c.CredhubCACert,
		CredhubURL:               c.CredhubURL,
		CredhubUsername:          c.CredhubUsername,
		ConcourseUsername:        c.ConcourseUsername,
		ConcoursePassword:        c.ConcoursePassword,
		GrafanaPassword:          c.GrafanaPassword,
		DirectorUsername:         c.DirectorUsername,
		DirectorPassword:         c.DirectorPassword,
		DirectorCACert:           c.DirectorCACert,
	}

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
		return bp, err
	}
	defer flyClient.Cleanup()

	concourseAlreadyRunning, err := flyClient.CanConnect()
	if err != nil {
		return bp, err
	}

	if !concourseAlreadyRunning {
		return bp, fmt.Errorf("In detach mode but it seems that concourse is not currently running")
	}

	// Allow a fly version discrepancy since we might be targetting an older Concourse
	if err = flyClient.SetDefaultPipeline(c, true); err != nil {
		return bp, err
	}

	bp, err = client.deployBosh(c, metadata, true)
	if err != nil {
		return bp, err
	}

	_, err = client.stdout.Write([]byte("\nUPGRADE RUNNING IN BACKGROUND\n\n"))

	return bp, err
}

// TerraformRequirements represents the required values for running terraform
type TerraformRequirements struct {
	Region                 string
	SourceAccessIP         string
	HostedZoneID           string
	HostedZoneRecordPrefix string
	Domain                 string
}

func (client *Client) checkPreTerraformConfigRequirements(conf config.Config, selfUpdate bool) (TerraformRequirements, error) {
	r := TerraformRequirements{
		Region:                 conf.Region,
		SourceAccessIP:         conf.SourceAccessIP,
		HostedZoneID:           conf.HostedZoneID,
		HostedZoneRecordPrefix: conf.HostedZoneRecordPrefix,
		Domain:                 conf.Domain,
	}

	region := client.provider.Region()
	if conf.Region != "" {
		if conf.Region != region {
			return r, fmt.Errorf("found previous deployment in %s. Refusing to deploy to %s as changing regions for existing deployments is not supported", conf.Region, region)
		}
	}

	r.Region = region

	// When in self-update mode do not override the user IP, since we already have access to the worker
	if !selfUpdate {
		var err error
		r.SourceAccessIP, err = client.setUserIP(conf)
		if err != nil {
			return r, err
		}
	}

	zone, err := client.setHostedZone(conf, conf.Domain)
	if err != nil {
		return r, err
	}
	r.HostedZoneID = zone.HostedZoneID
	r.HostedZoneRecordPrefix = zone.HostedZoneRecordPrefix
	r.Domain = zone.Domain

	return r, nil
}

// DirectorCerts represents the certificate of a Director
type DirectorCerts struct {
	DirectorCACert string
	DirectorCert   string
	DirectorKey    string
}

// Certs represents the certificate of a Concourse
type Certs struct {
	ConcourseCert             string
	ConcourseKey              string
	ConcourseUserProvidedCert bool
	ConcourseCACert           string
}

// Requirements represents the pre deployment requirements of a Concourse
type Requirements struct {
	Domain           string
	DirectorPublicIP string
	DirectorCerts    DirectorCerts
	Certs            Certs
}

func (client *Client) checkPreDeployConfigRequirements(c func(u *certs.User) (certs.AcmeClient, error), isDomainUpdated bool, cfg config.Config, metadata terraform.IAASMetadata) (Requirements, error) {
	cr := Requirements{
		Domain:           cfg.Domain,
		DirectorPublicIP: cfg.DirectorPublicIP,
	}

	if cfg.Domain == "" {
		domain, err := metadata.Get("ATCPublicIP")
		if err != nil {
			return cr, err
		}
		cr.Domain = domain
	}

	dc := DirectorCerts{
		DirectorCACert: cfg.DirectorCACert,
		DirectorCert:   cfg.DirectorCert,
		DirectorKey:    cfg.DirectorKey,
	}

	dc, err := client.ensureDirectorCerts(c, dc, cfg.Deployment, metadata)
	if err != nil {
		return cr, err
	}

	cr.DirectorCerts = dc

	cc := Certs{
		ConcourseCert:             cfg.ConcourseCert,
		ConcourseKey:              cfg.ConcourseKey,
		ConcourseUserProvidedCert: cfg.ConcourseUserProvidedCert,
		ConcourseCACert:           cfg.ConcourseCACert,
	}

	cc, err = client.ensureConcourseCerts(c, isDomainUpdated, cc, cfg.Deployment, cr.Domain)
	if err != nil {
		return cr, err
	}

	cr.Certs = cc

	cr.DirectorPublicIP, err = metadata.Get("DirectorPublicIP")
	if err != nil {
		return cr, err
	}

	return cr, nil
}

func (client *Client) ensureDirectorCerts(c func(u *certs.User) (certs.AcmeClient, error), dc DirectorCerts, deployment string, metadata terraform.IAASMetadata) (DirectorCerts, error) {
	// If we already have director certificates, don't regenerate as changing them will
	// force a bosh director re-deploy even if there are no other changes
	certs := dc
	if certs.DirectorCACert != "" {
		return certs, nil
	}

	ip, err := metadata.Get("DirectorPublicIP")
	if err != nil {
		return certs, err
	}
	_, err = client.stdout.Write(
		[]byte(fmt.Sprintf("\nGENERATING BOSH DIRECTOR CERTIFICATE (%s, 10.0.0.6)\n", ip)))
	if err != nil {
		return certs, err
	}

	directorCerts, err := client.certGenerator(c, deployment, ip, "10.0.0.6")
	if err != nil {
		return certs, err
	}

	certs.DirectorCACert = string(directorCerts.CACert)
	certs.DirectorCert = string(directorCerts.Cert)
	certs.DirectorKey = string(directorCerts.Key)

	return certs, nil
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

func (client *Client) ensureConcourseCerts(c func(u *certs.User) (certs.AcmeClient, error), domainUpdated bool, cc Certs, deployment, domain string) (Certs, error) {
	certs := cc

	if client.deployArgs.TLSCert != "" {
		certs.ConcourseCert = client.deployArgs.TLSCert
		certs.ConcourseKey = client.deployArgs.TLSKey
		certs.ConcourseUserProvidedCert = true

		return certs, nil
	}

	// Skip concourse re-deploy if certs have already been set,
	// unless domain has changed
	if certs.ConcourseCert != "" && !domainUpdated && timeTillExpiry(certs.ConcourseCert) > 28*24*time.Hour {
		return certs, nil
	}

	// If no domain has been provided by the user, the value of cfg.Domain is set to the ATC's public IP in checkPreDeployConfigRequirements
	Certs, err := client.certGenerator(c, deployment, domain)
	if err != nil {
		return certs, err
	}

	certs.ConcourseCert = string(Certs.Cert)
	certs.ConcourseKey = string(Certs.Key)
	certs.ConcourseCACert = string(Certs.CACert)

	return certs, nil
}

func (client *Client) deployBosh(config config.Config, metadata terraform.IAASMetadata, detach bool) (BoshParams, error) {
	bp := BoshParams{
		CredhubPassword:          config.CredhubPassword,
		CredhubAdminClientSecret: config.CredhubAdminClientSecret,
		CredhubCACert:            config.CredhubCACert,
		CredhubURL:               config.CredhubURL,
		CredhubUsername:          config.CredhubUsername,
		ConcourseUsername:        config.ConcourseUsername,
		ConcoursePassword:        config.ConcoursePassword,
		GrafanaPassword:          config.GrafanaPassword,
		DirectorUsername:         config.DirectorUsername,
		DirectorPassword:         config.DirectorPassword,
		DirectorCACert:           config.DirectorCACert,
	}

	boshClient, err := client.buildBoshClient(config, metadata)
	if err != nil {
		return bp, err
	}
	defer boshClient.Cleanup()

	boshStateBytes, err := loadDirectorState(client.configClient)
	if err != nil {
		return bp, err
	}
	boshCredsBytes, err := loadDirectorCreds(client.configClient)
	if err != nil {
		return bp, err
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
		return bp, err
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
		return bp, err
	}

	bp.CredhubPassword = cc.CredhubPassword
	bp.CredhubAdminClientSecret = cc.CredhubAdminClientSecret
	bp.CredhubCACert = cc.InternalTLS.CA
	bp.CredhubURL = fmt.Sprintf("https://%s:8844/", config.Domain)
	bp.CredhubUsername = "credhub-cli"
	bp.ConcourseUsername = "admin"
	if len(cc.AtcPassword) > 0 {
		bp.ConcoursePassword = cc.AtcPassword
		bp.GrafanaPassword = cc.AtcPassword
	}

	return bp, nil
}

func (client *Client) setUserIP(c config.Config) (string, error) {
	sourceAccessIP := c.SourceAccessIP
	userIP, err := client.ipChecker()
	if err != nil {
		return sourceAccessIP, err
	}

	if sourceAccessIP != userIP {
		sourceAccessIP = userIP
		_, err = client.stderr.Write([]byte(fmt.Sprintf(
			"\nWARNING: allowing access from local machine (address: %s)\n\n", userIP)))
		if err != nil {
			return sourceAccessIP, err
		}
	}

	return sourceAccessIP, nil
}

// HostedZone represents a DNS hosted zone
type HostedZone struct {
	HostedZoneID           string
	HostedZoneRecordPrefix string
	Domain                 string
}

func (client *Client) setHostedZone(c config.Config, domain string) (HostedZone, error) {
	zone := HostedZone{
		HostedZoneID:           c.HostedZoneID,
		HostedZoneRecordPrefix: c.HostedZoneRecordPrefix,
		Domain:                 c.Domain,
	}
	if domain == "" {
		return zone, nil
	}

	hostedZoneName, hostedZoneID, err := client.provider.FindLongestMatchingHostedZone(domain)
	if err != nil {
		return zone, err
	}
	zone.HostedZoneID = hostedZoneID
	zone.HostedZoneRecordPrefix = strings.TrimSuffix(domain, fmt.Sprintf(".%s", hostedZoneName))
	zone.Domain = domain

	_, err = client.stderr.Write([]byte(fmt.Sprintf(
		"\nWARNING: adding record %s to Route53 hosted zone %s ID: %s\n\n", domain, hostedZoneName, hostedZoneID)))
	if err != nil {
		return zone, err
	}
	return zone, err
}

const deployMsg = `DEPLOY SUCCESSFUL. Log in with:
fly --target {{.Project}} login{{if not .ConcourseUserProvidedCert}} --insecure{{end}} --concourse-url https://{{.Domain}} --username {{.ConcourseUsername}} --password {{.ConcoursePassword}}

Metrics available at https://{{.Domain}}:3000 using the same username and password

Log into credhub with:
eval "$(concourse-up info {{.Project}} --region {{.Region}} --env)"
`

func writeDeploySuccessMessage(config config.Config, metadata terraform.IAASMetadata, stdout io.Writer) error {
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

func eightRandomLetters() string {
	rand.Seed(time.Now().UTC().UnixNano())
	letterBytes := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
