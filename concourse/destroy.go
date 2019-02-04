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

	case awsConst: // nolint
		err = environment.Build(awsInputVarsMapFromConfig(conf))
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

	case gcpConst: // nolint
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
			"PublicCIDR":         conf.PublicCIDR,
			"PrivateCIDR":        conf.PrivateCIDR,
			"AllowIPs":           conf.AllowIPs,
			"ConfigBucket":       conf.ConfigBucket,
			"DBName":             conf.RDSDefaultDatabaseName,
			"DBPassword":         conf.RDSPassword,
			"DBTier":             conf.RDSInstanceClass,
			"DBUsername":         conf.RDSUsername,
			"Deployment":         conf.Deployment,
			"DNSManagedZoneName": conf.HostedZoneID,
			"DNSRecordSetPrefix": conf.HostedZoneRecordPrefix,
			"ExternalIP":         conf.SourceAccessIP,
			"GCPCredentialsJSON": credentialspath,
			"Namespace":          conf.Namespace,
			"Project":            project,
			"Region":             client.provider.Region(),
			"Tags":               "",
			"Zone":               zone,
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

	if client.provider.IAAS() == awsConst { // nolint
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
