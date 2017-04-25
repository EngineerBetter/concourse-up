package concourse

import (
	"io"

	"bitbucket.org/engineerbetter/concourse-up/bosh"
)

// Destroy destroys a concourse instance
func (client *Client) Destroy() error {
	conf, err := client.configClient.Load()
	if err != nil {
		return err
	}

	terraformClient, err := client.buildTerraformClient(conf)
	if err != nil {
		return err
	}
	defer terraformClient.Cleanup()

	metadata, err := terraformClient.Output()
	if err != nil {
		return err
	}

	boshClient, err := client.buildBoshClient(conf, metadata)
	if err != nil {
		return err
	}
	defer boshClient.Cleanup()

	boshStateBytes, err := loadDirectorState(client.configClient)
	if err != nil {
		return nil
	}

	boshStateBytes, err = boshClient.Delete(boshStateBytes)
	if err != nil {
		// If there was an error,
		// attempt to save the bosh state before propagating
		// This prevents issues when re-trying
		if boshStateBytes != nil {
			client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
		}
		return err
	}

	if err = client.configClient.DeleteAsset(bosh.StateFilename); err != nil {
		return err
	}

	if err := terraformClient.Destroy(); err != nil {
		return err
	}

	if err := client.configClient.DeleteAll(conf); err != nil {
		return err
	}

	return writeDestroySuccessMessage(client.stdout)
}

func writeDestroySuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nDESTROY SUCCESSFUL\n\n"))

	return err
}
