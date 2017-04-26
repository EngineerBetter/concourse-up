package concourse

import (
	"fmt"
	"io"

	"github.com/engineerbetter/concourse-up/bosh"
	"github.com/engineerbetter/concourse-up/config"
	"github.com/engineerbetter/concourse-up/terraform"
	"github.com/engineerbetter/concourse-up/util"
)

// Deploy deploys a concourse instance
func (client *Client) Deploy() error {
	config, err := client.loadConfigWithUserIP()
	if err != nil {
		return err
	}

	metadata, err := client.applyTerraform(config)
	if err != nil {
		return err
	}

	config, err = client.ensureCerts(config, metadata)
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

func (client *Client) ensureCerts(config *config.Config, metadata *terraform.Metadata) (*config.Config, error) {
	if config.DirectorCACert == "" {
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

		concourseCerts, err := client.certGenerator(config.Deployment, metadata.ELBDNSName.Value)
		if err != nil {
			return nil, err
		}

		config.ConcourseCert = string(concourseCerts.Cert)
		config.ConcourseKey = string(concourseCerts.Key)

		client.configClient.Update(config)
	}

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

func (client *Client) loadConfigWithUserIP() (*config.Config, error) {
	config, createdNewConfig, err := client.configClient.LoadOrCreate(client.args)
	if err != nil {
		return nil, err
	}

	if !createdNewConfig {
		if err = writeConfigLoadedSuccessMessage(client.stdout); err != nil {
			return nil, err
		}
	}

	userIP, err := util.FindUserIP()
	if err != nil {
		return nil, err
	}

	config.SourceAccessIP = userIP
	_, err = client.stderr.Write([]byte(fmt.Sprintf(
		"\nWARNING: allowing access from local machine (address: %s)\n\n", userIP)))
	if err != nil {
		return nil, err
	}

	return config, err
}

func writeDeploySuccessMessage(config *config.Config, metadata *terraform.Metadata, stdout io.Writer) error {
	_, err := stdout.Write([]byte(fmt.Sprintf(
		"\nDEPLOY SUCCESSFUL. Log in with:\n\nfly --target %s login --insecure --concourse-url https://%s --username %s --password %s\n\n",
		config.Project,
		metadata.ELBDNSName.Value,
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
