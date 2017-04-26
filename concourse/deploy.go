package concourse

import (
	"fmt"
	"io"

	"strings"

	"github.com/engineerbetter/concourse-up/bosh"
	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/terraform"
	"github.com/engineerbetter/concourse-up/util"
)

// Deploy deploys a concourse instance
func (client *Client) Deploy() error {
	config, err := client.loadConfig()
	if err != nil {
		return err
	}

	initialDomain := config.Domain

	if err = client.setUserIP(config); err != nil {
		return err
	}

	if err = client.setHostedZone(config); err != nil {
		return err
	}

	metadata, err := client.applyTerraform(config)
	if err != nil {
		return err
	}

	isDomainUpdated := initialDomain != config.Domain
	config, err = client.checkPredeployConfigRequiments(isDomainUpdated, config, metadata)
	if err != nil {
		return err
	}

	if err = client.deployBosh(config, metadata); err != nil {
		return err
	}

	if err = writeDeploySuccessMessage(config, metadata, client.stdout); err != nil {
		return err
	}

	return nil
}

func (client *Client) checkPredeployConfigRequiments(isDomainUpdated bool, config *config.Config, metadata *terraform.Metadata) (*config.Config, error) {
	config, err := client.ensureDomain(config, metadata)
	if err != nil {
		return nil, err
	}

	config, err = client.ensureDirectorCerts(config, metadata)
	if err != nil {
		return nil, err
	}

	config, err = client.ensureConcourseCerts(isDomainUpdated, config, metadata)
	if err != nil {
		return nil, err
	}

	if err := client.configClient.Update(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (client *Client) ensureDomain(config *config.Config, metadata *terraform.Metadata) (*config.Config, error) {
	if config.Domain == "" {
		config.Domain = metadata.ELBDNSName.Value
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
		[]byte(fmt.Sprintf("\nGENERATING CERTIFICATE FOR %s\n", ip)))
	if err != nil {
		return nil, err
	}

	directorCerts, err := client.certGenerator(config.Deployment, ip)
	if err != nil {
		return nil, err
	}

	config.DirectorCACert = string(directorCerts.CACert)
	config.DirectorCert = string(directorCerts.Cert)
	config.DirectorKey = string(directorCerts.Key)

	return config, nil
}

func (client *Client) ensureConcourseCerts(domainUpdated bool, config *config.Config, metadata *terraform.Metadata) (*config.Config, error) {
	if client.args["tls-cert"] != "" {
		config.ConcourseCert = client.args["tls-cert"]
		config.ConcourseKey = client.args["tls-key"]
		config.ConcourseUserProvidedCert = true

		return config, nil
	}

	// Skip concourse re-deploy if certs have already been set,
	// unless domain has changed
	if !(config.ConcourseCert == "" || domainUpdated) {
		return config, nil
	}

	concourseCerts, err := client.certGenerator(config.Deployment, config.Domain)
	if err != nil {
		return nil, err
	}

	config.ConcourseCert = string(concourseCerts.Cert)
	config.ConcourseKey = string(concourseCerts.Key)
	config.ConcourseCACert = string(concourseCerts.CACert)

	return config, nil
}

func (client *Client) applyTerraform(config *config.Config) (*terraform.Metadata, error) {
	terraformClient, err := client.buildTerraformClient(config)
	if err != nil {
		return nil, err
	}
	defer terraformClient.Cleanup()

	if err := terraformClient.Apply(); err != nil {
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

func (client *Client) deployBosh(config *config.Config, metadata *terraform.Metadata) error {
	boshClient, err := client.buildBoshClient(config, metadata)
	if err != nil {
		return err
	}
	defer boshClient.Cleanup()

	boshStateBytes, err := loadDirectorState(client.configClient)
	if err != nil {
		return nil
	}

	boshStateBytes, err = boshClient.Deploy(boshStateBytes)
	if err != nil {
		client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
		return err
	}

	return client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
}

func (client *Client) loadConfig() (*config.Config, error) {
	config, createdNewConfig, err := client.configClient.LoadOrCreate(client.args)
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
	domain := client.args["domain"]
	if domain == "" {
		return nil
	}

	hostedZoneName, hostedZoneID, err := client.hostedZoneFinder(domain)
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
		"\nDEPLOY SUCCESSFUL. Log in with:\n\nfly --target %s login%s --concourse-url https://%s --username %s --password %s\n\n",
		config.Project,
		flags,
		config.Domain,
		config.ConcourseUsername,
		config.ConcoursePassword,
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
