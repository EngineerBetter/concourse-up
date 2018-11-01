package concourse

import (
	"io"
)

// Destroy destroys a concourse instance
func (client *Client) Destroy() error {
	conf, err := client.configClient.Load()
	if err != nil {
		return err
	}

	environment, _, err := client.tfCLI.IAAS("GCP")
	if err != nil {
		return err
	}
	err = environment.Build(map[string]interface{}{
		"Zone":               conf.Region,
		"Tags":               "",
		"Project":            "concourse-up",
		"GCPCredentialsJSON": "/Users/eb/workspace/irbe-test/gcp-creds.json",
		"ExternalIP":         conf.SourceAccessIP,
		"Deployment":         conf.Deployment,
		"ConfigBucket":       conf.ConfigBucket,
	})

	// err = environment.Build(map[string]interface{}{
	// 	"AllowIPs":               conf.AllowIPs,
	// 	"AvailabilityZone":       conf.AvailabilityZone,
	// 	"ConfigBucket":           conf.ConfigBucket,
	// 	"Deployment":             conf.Deployment,
	// 	"HostedZoneID":           conf.HostedZoneID,
	// 	"HostedZoneRecordPrefix": conf.HostedZoneRecordPrefix,
	// 	"Namespace":              conf.Namespace,
	// 	"Project":                conf.Project,
	// 	"PublicKey":              conf.PublicKey,
	// 	"RDSDefaultDatabaseName": conf.RDSDefaultDatabaseName,
	// 	"RDSInstanceClass":       conf.RDSInstanceClass,
	// 	"RDSPassword":            conf.RDSPassword,
	// 	"RDSUsername":            conf.RDSUsername,
	// 	"Region":                 conf.Region,
	// 	"SourceAccessIP":         conf.SourceAccessIP,
	// 	"TFStatePath":            conf.TFStatePath,
	// 	"MultiAZRDS":             conf.MultiAZRDS,
	// })
	if err != nil {
		return err
	}
	// err = client.tfCLI.BuildOutput(environment, metadata)
	// if err != nil {
	// 	return err
	// }

	// vpcID, err := metadata.Get("VPCID")
	// if err != nil {
	// 	return err
	// }
	// volumesToDelete, err := client.provider.DeleteVMsInVPC(vpcID)
	// if err != nil {
	// 	return err
	// }

	err = client.tfCLI.Destroy(environment)
	if err != nil {
		return err
	}

	if err = client.configClient.DeleteAll(conf); err != nil {
		return err
	}

	// if err = client.provider.DeleteVolumes(volumesToDelete, iaas.DeleteVolume); err != nil {
	// 	return err
	// }

	return writeDestroySuccessMessage(client.stdout)
}
func writeDestroySuccessMessage(stdout io.Writer) error {
	_, err := stdout.Write([]byte("\nDESTROY SUCCESSFUL\n\n"))

	return err
}
