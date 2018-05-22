package concourse

import (
	"io"

	"github.com/EngineerBetter/concourse-up/iaas"
)

// Destroy destroys a concourse instance
func (client *Client) Destroy() error {
	conf, err := client.configClient.Load()
	if err != nil {
		return err
	}

	terraformClient, err := client.terraformClientFactory(client.iaasClient.IAAS(), conf, client.stdout, client.stderr)
	if err != nil {
		return err
	}
	defer terraformClient.Cleanup()

	metadata, err := terraformClient.Output()
	if err != nil {
		return err
	}

	volumesToDelete, err := client.iaasClient.DeleteVMsInVPC(metadata.VPCID.Value)
	if err != nil {
		return err
	}

	if err := terraformClient.Destroy(); err != nil {
		return err
	}

	if err := client.configClient.DeleteAll(conf); err != nil {
		return err
	}

	if err = client.iaasClient.DeleteVolumes(volumesToDelete, iaas.DeleteVolume, client.iaasClient.NewEC2Client); err != nil {
		return err
	}

	return writeDestroySuccessMessage(client.stdout)
}
func writeDestroySuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nDESTROY SUCCESSFUL\n\n"))

	return err
}
