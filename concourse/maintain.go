package concourse

import (
	"errors"
	"fmt"
	"time"

	"github.com/EngineerBetter/concourse-up/resource"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/commands/maintain"
)

type tasks struct {
	description string
	operation   string
	action      func(string, string) error
}

// Maintain fetches and builds the info
func (client *Client) Maintain(m maintain.Args) error {
	switch {
	case m.RenewNatsCertIsSet:
		return client.renewCert()
	}
	return nil
}

func (client *Client) renewCert() error {
	// waitForBOSHLocks will wait waitTime for BOSH to release its locks in order to proceed.
	// It will also printout a message to the user that the system is waiting for those locks.
	err := client.waitForBOSHLocks(10 * time.Minute)
	if err != nil {
		return err
	}

	// determineState will retrieve the last known state index and return 0 for an non existing one and n+1 for an existing one
	stateIndex, err := client.determineState()
	if err != nil {
		return err
	}

	tasks := []tasks{
		{"Adding new CA", resource.AddNewCa, client.createEnv},
		{"Recreating VMs for the first time", "first", client.recreate},
		{"Removing old CA", resource.RemoveOldCa, client.createEnv},
		{"Recreating VMs for the second time", "second", client.recreate},
		{"", "", client.cleanup},
		{"", "", client.deploy},
	}

	for i := stateIndex; i < len(tasks); i++ {
		fmt.Printf("current action: %s\n", tasks[i].description)
		err1 := tasks[i].action(tasks[i].description, tasks[i].operation)
		if err1 != nil {
			return err1
		}
	}

	return nil
}

func (client *Client) waitForBOSHLocks(waitTime time.Duration) error {
	return nil
}

func (client *Client) determineState() (int, error) {
	return 1, nil
}

func (client *Client) createEnv(description, operation string) error {
	c, err := client.configClient.Load()
	if err != nil {
		return err
	}

	environment, metadata, err := client.tfCLI.IAAS(client.provider.IAAS())
	if err != nil {
		return err
	}
	switch client.provider.IAAS() {
	case awsConst: // nolint
		err = environment.Build(map[string]interface{}{
			"AllowIPs":               c.AllowIPs,
			"AvailabilityZone":       c.AvailabilityZone,
			"ConfigBucket":           c.ConfigBucket,
			"Deployment":             c.Deployment,
			"HostedZoneID":           c.HostedZoneID,
			"HostedZoneRecordPrefix": c.HostedZoneRecordPrefix,
			"Namespace":              c.Namespace,
			"Project":                c.Project,
			"PublicKey":              c.PublicKey,
			"RDSDefaultDatabaseName": c.RDSDefaultDatabaseName,
			"RDSInstanceClass":       c.RDSInstanceClass,
			"RDSPassword":            c.RDSPassword,
			"RDSUsername":            c.RDSUsername,
			"Region":                 c.Region,
			"SourceAccessIP":         c.SourceAccessIP,
			"TFStatePath":            c.TFStatePath,
			"MultiAZRDS":             c.MultiAZRDS,
		})
		if err != nil {
			return err
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
		err1 = environment.Build(map[string]interface{}{
			"Region":             client.provider.Region(),
			"Zone":               client.provider.Zone(""),
			"Tags":               "",
			"Project":            project,
			"GCPCredentialsJSON": credentialspath,
			"ExternalIP":         c.SourceAccessIP,
			"Deployment":         c.Deployment,
			"ConfigBucket":       c.ConfigBucket,
			"DBTier":             c.RDSInstanceClass,
			"DBPassword":         c.RDSPassword,
			"DBUsername":         c.RDSUsername,
			"DBName":             c.RDSDefaultDatabaseName,
			"AllowIPs":           c.AllowIPs,
			"DNSManagedZoneName": c.HostedZoneID,
			"DNSRecordSetPrefix": c.HostedZoneRecordPrefix,
		})
		if err1 != nil {
			return err1
		}
	default:
		return errors.New("concourse:deploy:unsupported iaas " + client.deployArgs.IAAS)
	}

	err = client.tfCLI.BuildOutput(environment, metadata)
	if err != nil {
		return err
	}

	boshClient, err := client.buildBoshClient(c, metadata)
	if err != nil {
		return err
	}
	defer boshClient.Cleanup()

	boshStateBytes, err := loadDirectorState(client.configClient)
	if err != nil {
		return err
	}

	boshCredsBytes, err := loadDirectorCreds(client.configClient)
	if err != nil {
		return err
	}

	boshStateBytes, boshCredsBytes, err = boshClient.CreateEnv(boshStateBytes, boshCredsBytes, operation)
	err1 := client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
	if err == nil {
		err = err1
	}
	err1 = client.configClient.StoreAsset(bosh.CredsFilename, boshCredsBytes)
	if err == nil {
		err = err1
	}
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) recreate(description, operation string) error {
	c, err := client.configClient.Load()
	if err != nil {
		return err
	}

	environment, metadata, err := client.tfCLI.IAAS(client.provider.IAAS())
	if err != nil {
		return err
	}
	switch client.provider.IAAS() {
	case awsConst: // nolint
		err = environment.Build(map[string]interface{}{
			"AllowIPs":               c.AllowIPs,
			"AvailabilityZone":       c.AvailabilityZone,
			"ConfigBucket":           c.ConfigBucket,
			"Deployment":             c.Deployment,
			"HostedZoneID":           c.HostedZoneID,
			"HostedZoneRecordPrefix": c.HostedZoneRecordPrefix,
			"Namespace":              c.Namespace,
			"Project":                c.Project,
			"PublicKey":              c.PublicKey,
			"RDSDefaultDatabaseName": c.RDSDefaultDatabaseName,
			"RDSInstanceClass":       c.RDSInstanceClass,
			"RDSPassword":            c.RDSPassword,
			"RDSUsername":            c.RDSUsername,
			"Region":                 c.Region,
			"SourceAccessIP":         c.SourceAccessIP,
			"TFStatePath":            c.TFStatePath,
			"MultiAZRDS":             c.MultiAZRDS,
		})
		if err != nil {
			return err
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
		err1 = environment.Build(map[string]interface{}{
			"Region":             client.provider.Region(),
			"Zone":               client.provider.Zone(""),
			"Tags":               "",
			"Project":            project,
			"GCPCredentialsJSON": credentialspath,
			"ExternalIP":         c.SourceAccessIP,
			"Deployment":         c.Deployment,
			"ConfigBucket":       c.ConfigBucket,
			"DBTier":             c.RDSInstanceClass,
			"DBPassword":         c.RDSPassword,
			"DBUsername":         c.RDSUsername,
			"DBName":             c.RDSDefaultDatabaseName,
			"AllowIPs":           c.AllowIPs,
			"DNSManagedZoneName": c.HostedZoneID,
			"DNSRecordSetPrefix": c.HostedZoneRecordPrefix,
		})
		if err1 != nil {
			return err1
		}
	default:
		return errors.New("concourse:deploy:unsupported iaas " + client.deployArgs.IAAS)
	}

	err = client.tfCLI.BuildOutput(environment, metadata)
	if err != nil {
		return err
	}

	boshClient, err := client.buildBoshClient(c, metadata)
	if err != nil {
		return err
	}
	defer boshClient.Cleanup()

	err = boshClient.Recreate()
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) deploy(description, operation string) error {
	fmt.Printf("deploy %s\n", description)
	return nil
}

func (client *Client) cleanup(description, operation string) error {
	fmt.Printf("cleanup %s\n", description)
	return nil
}
