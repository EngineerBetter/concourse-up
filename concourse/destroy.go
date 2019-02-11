package concourse

import (
	"fmt"
	"io"

	"github.com/EngineerBetter/concourse-up/iaas"
)

// Destroy destroys a concourse instance
func (client *Client) Destroy() error {

	conf, err := client.configClient.Load()
	if err != nil {
		return err
	}

	tfOutputs, err := client.tfCLI.OutputsFor(client.provider.IAAS())
	if err != nil {
		return err
	}

	tfInputVars := client.tfInputVarsFactory.NewInputVars(conf)

	var volumesToDelete []string

	switch client.provider.IAAS() {

	case iaas.AWS: // nolint
		err = client.tfCLI.BuildOutput(tfInputVars, tfOutputs)
		if err != nil {
			return err
		}
		vpcID, err1 := tfOutputs.Get("VPCID")
		if err1 != nil {
			return err1
		}
		volumesToDelete, err1 = client.provider.DeleteVMsInVPC(vpcID)
		if err1 != nil {
			return err1
		}

	case iaas.GCP: // nolint
		project, err1 := client.provider.Attr("project")
		if err1 != nil {
			return err1
		}
		zone := client.provider.Zone("")
		err1 = client.provider.DeleteVMsInDeployment(zone, project, conf.Deployment)
		if err1 != nil {
			return err1
		}
	}

	err = client.tfCLI.Destroy(tfInputVars)
	if err != nil {
		return err
	}

	if client.provider.IAAS() == iaas.AWS { // nolint
		if len(volumesToDelete) > 0 {
			fmt.Printf("Scheduling to delete %v volumes\n", len(volumesToDelete))
		}
		if err1 := client.provider.DeleteVolumes(volumesToDelete, iaas.DeleteVolume); err1 != nil {
			return err1
		}
	}

	if err = client.configClient.DeleteAll(conf); err != nil {
		return err
	}

	return writeDestroySuccessMessage(client.stdout)
}
func writeDestroySuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nDESTROY SUCCESSFUL\n\n"))

	return err
}
