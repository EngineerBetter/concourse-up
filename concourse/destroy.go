package concourse

import (
	"fmt"
	"io"
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

	if err := boshClient.DeleteDirector(); err != nil {
		if err := writeDeleteBoshDirectorErrorWarning(client.stderr, err.Error()); err != nil {
			return err
		}
	}

	if err := terraformClient.Destroy(); err != nil {
		return err
	}

	return writeDestroySuccessMessage(client.stdout)
}

func writeDeleteBoshDirectorErrorWarning(stderr io.Writer, message string) error {
	_, err := stderr.Write([]byte(fmt.Sprintf(
		"Warning error deleting bosh director. Continuing with terraform deletion.\n\t%s", message)))

	return err
}
func writeDestroySuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nDESTROY SUCCESSFUL\n\n"))

	return err
}
