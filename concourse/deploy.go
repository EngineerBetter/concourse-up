package concourse

import (
	"fmt"
	"io"

	"strings"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
	"bitbucket.org/engineerbetter/concourse-up/config"
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

		certs, err := client.certGenerator(config.Deployment, ip)
		if err != nil {
			return nil, err
		}

		config.DirectorCACert = string(certs.CACert)
		config.DirectorCert = string(certs.Cert)
		config.DirectorKey = string(certs.Key)

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

	boshStateBytes, err := boshClient.Deploy()
	if err != nil {
		return err
	}

	return client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
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

func writeDeploySuccessMessage(config *config.Config, metadata *terraform.Metadata, stdout io.Writer) error {
	caCert := strings.Replace(config.DirectorCACert, "\n", "\n\t\t", -1)

	_, err := stdout.Write([]byte(fmt.Sprintf(
		"\nDEPLOY SUCCESSFUL. Bosh connection credentials:\n\tIP Address: %s\n\tUsername: %s\n\tPassword: %s\n\tCA Cert:\n\t\t%s\n\n",
		metadata.DirectorPublicIP.Value,
		config.DirectorUsername,
		config.DirectorPassword,
		caCert,
	)))

	return err
}

func writeConfigLoadedSuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nUSING PREVIOUS DEPLOYMENT CONFIG\n"))

	return err
}
