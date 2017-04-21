package concourse

import (
	"fmt"
	"io"

	"bitbucket.org/engineerbetter/concourse-up/config"
	"bitbucket.org/engineerbetter/concourse-up/director"
	"bitbucket.org/engineerbetter/concourse-up/terraform"
	"bitbucket.org/engineerbetter/concourse-up/util"
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

	if err = client.deployBosh(config, metadata); err != nil {
		return err
	}

	if err = writeDeploySuccessMessage(config, metadata, client.stdout); err != nil {
		return err
	}

	return nil
}

func (client *Client) applyTerraform(config *config.Config) (metadata *terraform.Metadata, err error) {
	var terraformClient terraform.IClient
	terraformClient, err = client.buildTerraformClient(config)
	if err != nil {
		return
	}
	defer func() { err = terraformClient.Cleanup() }()

	err = terraformClient.Apply()
	if err != nil {
		return
	}

	metadata, err = terraformClient.Output()
	return
}

func (client *Client) deployBosh(config *config.Config, metadata *terraform.Metadata) (err error) {
	var boshInitClient director.IBoshInitClient
	boshInitClient, err = client.buildBoshInitClient(config, metadata)
	if err != nil {
		return
	}
	defer func() { err = boshInitClient.Cleanup() }()

	var boshStateBytes []byte
	boshStateBytes, err = boshInitClient.Deploy()
	if err != nil {
		return
	}

	err = client.configClient.StoreAsset(director.StateFilename, boshStateBytes)

	return
}

func (client *Client) loadConfigWithUserIP() (*config.Config, error) {
	config, createdNewConfig, err := client.configClient.LoadOrCreate()
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

func loadDirectorState(configClient config.IClient) ([]byte, error) {
	hasState, err := configClient.HasAsset(director.StateFilename)
	if err != nil {
		return nil, err
	}

	if !hasState {
		return nil, nil
	}

	return configClient.LoadAsset(director.StateFilename)
}

func writeDeploySuccessMessage(config *config.Config, metadata *terraform.Metadata, stdout io.Writer) error {
	_, err := stdout.Write([]byte(fmt.Sprintf(
		"\nDEPLOY SUCCESSFUL. Bosh connection credentials:\n\tIP Address: %s\n\tUsername: %s\n\tPassword: %s\n\n",
		metadata.DirectorPublicIP.Value,
		config.DirectorUsername,
		config.DirectorPassword,
	)))

	return err
}

func writeConfigLoadedSuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nUSING PREVIOUS DEPLOYMENT CONFIG\n"))

	return err
}
