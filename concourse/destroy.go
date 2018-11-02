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

	environment, metadata, err := client.tfCLI.IAAS(client.provider.IAAS())
	if err != nil {
		return err
	}

	switch client.provider.IAAS() {
	case "AWS":
		err = environment.Build(map[string]interface{}{
			"AllowIPs":               conf.AllowIPs,
			"AvailabilityZone":       conf.AvailabilityZone,
			"ConfigBucket":           conf.ConfigBucket,
			"Deployment":             conf.Deployment,
			"HostedZoneID":           conf.HostedZoneID,
			"HostedZoneRecordPrefix": conf.HostedZoneRecordPrefix,
			"Namespace":              conf.Namespace,
			"Project":                conf.Project,
			"PublicKey":              conf.PublicKey,
			"RDSDefaultDatabaseName": conf.RDSDefaultDatabaseName,
			"RDSInstanceClass":       conf.RDSInstanceClass,
			"RDSPassword":            conf.RDSPassword,
			"RDSUsername":            conf.RDSUsername,
			"Region":                 conf.Region,
			"SourceAccessIP":         conf.SourceAccessIP,
			"TFStatePath":            conf.TFStatePath,
			"MultiAZRDS":             conf.MultiAZRDS,
		})
		if err != nil {
			return err
		}
		err = client.tfCLI.BuildOutput(environment, metadata)
		if err != nil {
			return err
		}

		vpcID, err := metadata.Get("VPCID")
		if err != nil {
			return err
		}
		volumesToDelete, err := client.provider.DeleteVMsInVPC(vpcID)
		if err != nil {
			return err
		}

		if err = client.provider.DeleteVolumes(volumesToDelete, iaas.DeleteVolume); err != nil {
			return err
		}

	case "GCP":
		project, err := client.provider.Attr("project")
		if err != nil {
			return err
		}
		credentialspath, err := client.provider.Attr("path")
		if err != nil {
			return err
		}
		err = environment.Build(map[string]interface{}{
			"Region":             client.provider.Region(),
			"Zone":               client.provider.Zone(),
			"Tags":               "",
			"Project":            project,
			"GCPCredentialsJSON": credentialspath,
			"ExternalIP":         conf.SourceAccessIP,
			"Deployment":         conf.Deployment,
			"ConfigBucket":       conf.ConfigBucket,
		})
		if err != nil {
			return err
		}
	}

	err = client.tfCLI.Destroy(environment)
	if err != nil {
		return err
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
