package concourse

import (
	"fmt"
	"io"

	"strings"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/fly"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/EngineerBetter/concourse-up/util"
)

// Deploy deploys a concourse instance
func (client *Client) Deploy() error {
	config, err := client.loadConfig()
	if err != nil {
		return err
	}

	initialDomain := config.Domain

	config, err = client.checkPreTerraformConfigRequirements(config)
	if err != nil {
		return err
	}

	metadata, err := client.applyTerraform(config)
	if err != nil {
		return err
	}

	isDomainUpdated := initialDomain != config.Domain
	config, err = client.checkPreDeployConfigRequiments(isDomainUpdated, config, metadata)
	if err != nil {
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
	)
	if err != nil {
		return err
	}
	defer flyClient.Cleanup()

	if client.deployArgs.SelfUpdate {
		return client.updateBoshAndPipeline(config, metadata, flyClient)
	}

	return client.deployBoshAndPipeline(config, metadata, flyClient)
}

func (client *Client) deployBoshAndPipeline(config *config.Config, metadata *terraform.Metadata, flyClient fly.IClient) error {
	// When we are deploying for the first time rather than updating
	// ensure that the pipeline is set _after_ the concourse is deployed
	if err := client.deployBosh(config, metadata, false); err != nil {
		return err
	}

	if err := flyClient.SetDefaultPipeline(client.deployArgs, config, false); err != nil {
		return err
	}

	if err := writeDeploySuccessMessage(config, metadata, client.stdout); err != nil {
		return err
	}

	return nil
}

func (client *Client) updateBoshAndPipeline(config *config.Config, metadata *terraform.Metadata, flyClient fly.IClient) error {
	// If concourse is already running this is an update rather than a fresh deploy
	// When updating we need to deploy the BOSH as the final step in order to
	// Detach from the update, so the update job can exit
	concourseAlreadyRunning, err := flyClient.CanConnect()
	if err != nil {
		return err
	}

	if !concourseAlreadyRunning {
		return fmt.Errorf("In detach mode but it seems that concourse is not currently running")
	}

	// Allow a fly version discrepancy since we might be targetting an older Concourse
	if err = flyClient.SetDefaultPipeline(client.deployArgs, config, true); err != nil {
		return err
	}

	if err = client.deployBosh(config, metadata, true); err != nil {
		return err
	}

	_, err = client.stdout.Write([]byte("\nUPGRADE RUNNING IN BACKGROUND\n\n"))

	return err
}

func (client *Client) checkPreTerraformConfigRequirements(conf *config.Config) (*config.Config, error) {
	region := client.deployArgs.AWSRegion

	if conf.Region != "" {
		if conf.Region != region {
			return nil, fmt.Errorf("found previous deployment in %s. Refusing to deploy to %s as changing regions for existing deployments is not supported", conf.Region, region)
		}
	}

	conf.Region = region

	// If the RDS instance size has manually set, override the existing size in the config
	if client.deployArgs.DBSizeIsSet {
		conf.RDSInstanceClass = config.DBSizes[client.deployArgs.DBSize]
	}

	// When in self-update mode do not override the user IP, since we already have access to the worker
	if !client.deployArgs.SelfUpdate {
		if err := client.setUserIP(conf); err != nil {
			return nil, err
		}
	}

	if err := client.setHostedZone(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func (client *Client) checkPreDeployConfigRequiments(isDomainUpdated bool, config *config.Config, metadata *terraform.Metadata) (*config.Config, error) {
	if client.deployArgs.Domain == "" {
		config.Domain = metadata.ATCPublicIP.Value
	}

	config, err := client.ensureDirectorCerts(config, metadata)
	if err != nil {
		return nil, err
	}

	config, err = client.ensureConcourseCerts(isDomainUpdated, config, metadata)
	if err != nil {
		return nil, err
	}

	config.ConcourseWorkerCount = client.deployArgs.WorkerCount
	config.ConcourseWorkerSize = client.deployArgs.WorkerSize
	config.DirectorPublicIP = metadata.DirectorPublicIP.Value

	if err := client.configClient.Update(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (client *Client) ensureDirectorCerts(config *config.Config, metadata *terraform.Metadata) (*config.Config, error) {
	// If we already have director certificates, don't regenerate as changing them will
	// force a bosh director re-deploy even if there are no other changes
	if config.DirectorCACert != "" {
		return config, nil
	}

	ip := metadata.DirectorPublicIP.Value
	_, err := client.stdout.Write(
		[]byte(fmt.Sprintf("\nGENERATING BOSH DIRECTOR CERTIFICATE (%s, 10.0.0.6)\n", ip)))
	if err != nil {
		return nil, err
	}

	directorCerts, err := client.certGenerator(config.Deployment, ip, "10.0.0.6")
	if err != nil {
		return nil, err
	}

	config.DirectorCACert = string(directorCerts.CACert)
	config.DirectorCert = string(directorCerts.Cert)
	config.DirectorKey = string(directorCerts.Key)

	return config, nil
}

func (client *Client) ensureConcourseCerts(domainUpdated bool, config *config.Config, metadata *terraform.Metadata) (*config.Config, error) {
	if client.deployArgs.TLSCert != "" {
		config.ConcourseCert = client.deployArgs.TLSCert
		config.ConcourseKey = client.deployArgs.TLSKey
		config.ConcourseUserProvidedCert = true

		return config, nil
	}

	// Skip concourse re-deploy if certs have already been set,
	// unless domain has changed
	if !(config.ConcourseCert == "" || domainUpdated) {
		return config, nil
	}

	sans := []string{config.Domain}
	if config.Domain != metadata.ATCPublicIP.Value {
		sans = append(sans, metadata.ATCPublicIP.Value)
	}
	concourseCerts, err := client.certGenerator(config.Deployment, sans...)
	if err != nil {
		return nil, err
	}

	config.ConcourseCert = string(concourseCerts.Cert)
	config.ConcourseKey = string(concourseCerts.Key)
	config.ConcourseCACert = string(concourseCerts.CACert)

	return config, nil
}

func (client *Client) applyTerraform(config *config.Config) (*terraform.Metadata, error) {
	terraformClient, err := client.terraformClientFactory(client.iaasClient.IAAS(), config, client.stdout, client.stderr)
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

func (client *Client) deployBosh(config *config.Config, metadata *terraform.Metadata, detach bool) error {
	boshClient, err := client.buildBoshClient(config, metadata)
	if err != nil {
		return err
	}
	defer boshClient.Cleanup()

	boshStateBytes, err := loadDirectorState(client.configClient)
	if err != nil {
		return nil
	}

	boshStateBytes, err = boshClient.Deploy(boshStateBytes, detach)
	if err != nil {
		client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
		return err
	}

	return client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
}

func (client *Client) loadConfig() (*config.Config, error) {
	config, createdNewConfig, err := client.configClient.LoadOrCreate(client.deployArgs)
	if err != nil {
		return nil, err
	}

	if !createdNewConfig {
		if err = writeConfigLoadedSuccessMessage(client.stdout); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (client *Client) setUserIP(config *config.Config) error {
	userIP, err := util.FindUserIP()
	if err != nil {
		return err
	}

	if config.SourceAccessIP != userIP {
		config.SourceAccessIP = userIP
		_, err = client.stderr.Write([]byte(fmt.Sprintf(
			"\nWARNING: allowing access from local machine (address: %s)\n\n", userIP)))
		if err != nil {
			return err
		}
		if err = client.configClient.Update(config); err != nil {
			return err
		}
	}

	return nil
}

func (client *Client) setHostedZone(config *config.Config) error {
	domain := client.deployArgs.Domain
	if client.deployArgs.Domain == "" {
		return nil
	}

	hostedZoneName, hostedZoneID, err := client.iaasClient.FindLongestMatchingHostedZone(domain)
	if err != nil {
		return err
	}
	config.HostedZoneID = hostedZoneID
	config.HostedZoneRecordPrefix = strings.TrimSuffix(domain, fmt.Sprintf(".%s", hostedZoneName))
	config.Domain = domain

	_, err = client.stderr.Write([]byte(fmt.Sprintf(
		"\nWARNING: adding record %s to Route53 hosted zone %s ID: %s\n\n", domain, hostedZoneName, hostedZoneID)))
	if err != nil {
		return err
	}
	if err = client.configClient.Update(config); err != nil {
		return err
	}

	return nil
}

func writeDeploySuccessMessage(config *config.Config, metadata *terraform.Metadata, stdout io.Writer) error {
	flags := ""
	if !config.ConcourseUserProvidedCert {
		flags = " --insecure"
	}
	_, err := stdout.Write([]byte(fmt.Sprintf(
		"\nDEPLOY SUCCESSFUL. Log in with:\n\nfly --target %s login%s --concourse-url https://%s --username %s --password %s\n\nMetrics available at https://%s:3000 using the same username and password\n\n",
		config.Project,
		flags,
		config.Domain,
		config.ConcourseUsername,
		config.ConcoursePassword,
		config.Domain,
	)))

	return err
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
