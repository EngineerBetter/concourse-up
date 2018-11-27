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

	environment, metadata, err := client.tfCLI.IAAS(client.provider.IAAS())
	if err != nil {
		return err
	}

	var volumesToDelete []string

	switch client.provider.IAAS() {

	case "AWS": // nolint
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh_%s", eightRandomLetters())

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
		vpcID, err1 := metadata.Get("VPCID")
		if err1 != nil {
			return err1
		}
		volumesToDelete, err1 = client.provider.DeleteVMsInVPC(vpcID)
		if err1 != nil {
			return err1
		}

	case "GCP": // nolint
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh-%s", eightRandomLetters())

		project, err1 := client.provider.Attr("project")
		if err1 != nil {
			return err1
		}
		credentialspath, err1 := client.provider.Attr("credentials_path")
		if err1 != nil {
			return err1
		}
		zone := client.provider.Zone("")
		err1 = environment.Build(map[string]interface{}{
			"Region":             client.provider.Region(),
			"Zone":               zone,
			"Tags":               "",
			"Project":            project,
			"GCPCredentialsJSON": credentialspath,
			"ExternalIP":         conf.SourceAccessIP,
			"Deployment":         conf.Deployment,
			"ConfigBucket":       conf.ConfigBucket,
			"DBName":             conf.RDSDefaultDatabaseName,
			"DBTier":             "db-f1-micro",
			"DBPassword":         conf.RDSPassword,
			"DBUsername":         conf.RDSUsername,
		})
		if err1 != nil {
			return nil
		}
		err1 = client.provider.DeleteVMsInDeployment(zone, project, conf.Deployment)
		if err1 != nil {
			return err1
		}
	}

	err = client.tfCLI.Destroy(environment)
	if err != nil {
		return err
	}

	if client.provider.IAAS() == "AWS" { // nolint
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
