package concourse

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/EngineerBetter/concourse-up/resource"
	"github.com/EngineerBetter/concourse-up/util/yaml"

	"github.com/EngineerBetter/concourse-up/bosh"
	"github.com/EngineerBetter/concourse-up/commands/maintain"
)

type tasks struct {
	description string
	operation   string
	action      func(string, string) error
}

// Maintenance is a struct representing values used by the maintenance command
type Maintenance struct {
	StatusIndex int `json:"status_index"`
}

// Tables represents the output of bosh locks
type Tables struct {
	Tables []Table `json:"Tables"`
}

// Table represents the subelements of the Tables struct
type Table struct {
	Content string
	Rows    []interface{} `json:"Rows"`
}

const maintenanceFilename = "maintenance.json"

// Maintain fetches and builds the info
func (client *Client) Maintain(m maintain.Args) error {
	switch {
	case m.RenewNatsCertIsSet:
		return client.renewCert(m)
	}
	return nil
}

func (client *Client) renewCert(m maintain.Args) error {

	_ = client.waitForBOSHLocks(10 * time.Minute)

	maintenance, err := client.retrieveStage()
	if err != nil {
		return err
	}

	var stageIndex int
	if m.StageIsSet {
		stageIndex = m.Stage
	} else {
		stageIndex, err = client.determineStage(maintenance)
		if err != nil {
			return err
		}
	}

	tasks := []tasks{
		{"Adding new CA", resource.AddNewCa, client.createEnv},
		{"Recreating VMs for the first time", "first", client.recreate},
		{"Removing old CA", resource.RemoveOldCa, client.createEnv},
		{"Recreating VMs for the second time", "second", client.recreate},
		{"Cleaning up director-creds.yml", "", client.cleanup},
	}

	if stageIndex >= len(tasks) {
		return fmt.Errorf("Invalid stage index")
	}

	for i := stageIndex; i < len(tasks); i++ {
		fmt.Printf("current action: %s\n", tasks[i].description)
		err1 := tasks[i].action(tasks[i].description, tasks[i].operation)
		if err1 != nil {
			return err1
		}
		err1 = client.updateStage(i, maintenance)
		if err1 != nil {
			return err1
		}
	}
	err = client.updateStage(-1, maintenance)
	return err
}

// constructBoshClient creates a boshClient for use in this package
func (client *Client) constructBoshClient() (*bosh.IClient, error) {
	conf, err := client.configClient.Load()
	if err != nil {
		return nil, err
	}

	tfOutputs, err := client.tfCLI.OutputsFor(client.provider.IAAS())
	if err != nil {
		return nil, err
	}

	tfInputVars := client.tfInputVarsFactory.NewInputVars(conf)

	err = client.tfCLI.BuildOutput(tfInputVars, tfOutputs)
	if err != nil {
		return nil, err
	}

	boshClient, err := client.buildBoshClient(conf, tfOutputs)
	if err != nil {
		return nil, err
	}
	return &boshClient, nil
}

// checkIfLocked checks if the lock is taken on the director
// returns true if the lock is taken
func (client *Client) checkIfLocked() (bool, error) {
	var tables Tables
	boshClientPointer, err := client.constructBoshClient()
	if err != nil {
		return true, err
	}
	boshClient := *boshClientPointer
	defer boshClient.Cleanup()
	lockBytes, err := boshClient.Locks()
	if err != nil {
		return true, err
	}
	err = json.Unmarshal(lockBytes, &tables)
	if err != nil {
		return true, err
	}
	for _, val := range tables.Tables {
		if val.Content == "locks" {
			return len(val.Rows) != 0, nil
		}
	}
	return true, nil
}

// waitForBOSHLocks will wait waitTime for BOSH to release its locks in order to proceed.
// It will also printout a message to the user that the system is waiting for those locks.
func (client *Client) waitForBOSHLocks(waitTime time.Duration) error {
	start := time.Now().UTC()
	for {
		fmt.Println("Waiting for BOSH lock to become available")
		locked, err := client.checkIfLocked()
		if err != nil {
			return err
		}
		if !locked {
			return nil
		}
		if time.Since(start) > waitTime {
			return fmt.Errorf("BOSH lock failed to become available after %ds", waitTime/1000)
		}
	}
}

// retrieveStage will retrieve the maintenance object from the config bucket
// if the object is not found it will create one with statusIndex of -1
func (client *Client) retrieveStage() (*Maintenance, error) {
	var maintenance Maintenance
	fileExists, err := client.configClient.HasAsset(maintenanceFilename)
	if err != nil {
		return nil, err
	}
	if fileExists {
		fileContents, err := client.configClient.LoadAsset(maintenanceFilename)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(fileContents, &maintenance)
		if err != nil {
			return nil, err
		}
	} else {
		maintenance = Maintenance{StatusIndex: -1}
	}
	return &maintenance, nil
}

// determineStage returns the index of the next operation to be run
func (client *Client) determineStage(maintenance *Maintenance) (int, error) {
	return maintenance.StatusIndex + 1, nil
}

// updateStage stores the specified index in the maintenance object in the config bucket
func (client *Client) updateStage(index int, maintenance *Maintenance) error {
	maintenance.StatusIndex = index
	maintenanceBytes, err := json.Marshal(maintenance)
	if err != nil {
		return err
	}
	return client.configClient.StoreAsset(maintenanceFilename, maintenanceBytes)
}

// createEnv runs bosh create-env
func (client *Client) createEnv(description, operation string) error {
	boshClientPointer, err := client.constructBoshClient()
	if err != nil {
		return err
	}
	boshClient := *boshClientPointer
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

// recreate runs bosh recreate
func (client *Client) recreate(description, operation string) error {
	boshClientPointer, err := client.constructBoshClient()
	if err != nil {
		return err
	}
	boshClient := *boshClientPointer
	defer boshClient.Cleanup()

	err = boshClient.Recreate()
	if err != nil {
		return err
	}
	return nil
}

// cleanup cleans up the director-creds.yml file
func (client *Client) cleanup(description, operation string) error {
	directorCredsBytes, err := loadDirectorCreds(client.configClient)
	if err != nil {
		return err
	}
	natsCACA, err := yaml.Path(directorCredsBytes, "nats_ca_2/ca")
	if err != nil {
		return err
	}
	natsCACert, err := yaml.Path(directorCredsBytes, "nats_ca_2/certificate")
	if err != nil {
		return err
	}
	natsCAKey, err := yaml.Path(directorCredsBytes, "nats_ca_2/private_key")
	if err != nil {
		return err
	}
	natsClientsDirectorTLSCA, err := yaml.Path(directorCredsBytes, "nats_clients_director_tls_2/ca")
	if err != nil {
		return err
	}
	natsClientsDirectorTLSCert, err := yaml.Path(directorCredsBytes, "nats_clients_director_tls_2/certificate")
	if err != nil {
		return err
	}
	natsClientsDirectorTLSKey, err := yaml.Path(directorCredsBytes, "nats_clients_director_tls_2/private_key")
	if err != nil {
		return err
	}
	natsClientsHealthMonitorTLSCA, err := yaml.Path(directorCredsBytes, "nats_clients_health_monitor_tls_2/ca")
	if err != nil {
		return err
	}
	natsClientsHealthMonitorTLSCert, err := yaml.Path(directorCredsBytes, "nats_clients_health_monitor_tls_2/certificate")
	if err != nil {
		return err
	}
	natsClientsHealthMonitorTLSKey, err := yaml.Path(directorCredsBytes, "nats_clients_health_monitor_tls_2/private_key")
	if err != nil {
		return err
	}
	natsServerTLSCA, err := yaml.Path(directorCredsBytes, "nats_server_tls_2/ca")
	if err != nil {
		return err
	}
	natsServerTLSCert, err := yaml.Path(directorCredsBytes, "nats_server_tls_2/certificate")
	if err != nil {
		return err
	}
	natsServerTLSKey, err := yaml.Path(directorCredsBytes, "nats_server_tls_2/private_key")
	if err != nil {
		return err
	}

	var re = regexp.MustCompile(`\|\n| {2,}`)

	correctedCreds, err := yaml.Interpolate(string(directorCredsBytes), resource.CleanupCerts, map[string]interface{}{
		"nats_ca_ca":                                  re.ReplaceAllString(natsCACA, ""),
		"nats_ca_certificate":                         re.ReplaceAllString(natsCACert, ""),
		"nats_ca_private_key":                         re.ReplaceAllString(natsCAKey, ""),
		"nats_clients_director_tls_ca":                re.ReplaceAllString(natsClientsDirectorTLSCA, ""),
		"nats_clients_director_tls_certificate":       re.ReplaceAllString(natsClientsDirectorTLSCert, ""),
		"nats_clients_director_tls_private_key":       re.ReplaceAllString(natsClientsDirectorTLSKey, ""),
		"nats_clients_health_monitor_tls_ca":          re.ReplaceAllString(natsClientsHealthMonitorTLSCA, ""),
		"nats_clients_health_monitor_tls_certificate": re.ReplaceAllString(natsClientsHealthMonitorTLSCert, ""),
		"nats_clients_health_monitor_tls_private_key": re.ReplaceAllString(natsClientsHealthMonitorTLSKey, ""),
		"nats_server_tls_ca":                          re.ReplaceAllString(natsServerTLSCA, ""),
		"nats_server_tls_certificate":                 re.ReplaceAllString(natsServerTLSCert, ""),
		"nats_server_tls_private_key":                 re.ReplaceAllString(natsServerTLSKey, ""),
	})
	if err != nil {
		return err
	}
	err = client.configClient.StoreAsset("director-creds-backup.yml", directorCredsBytes)
	if err != nil {
		return err
	}
	err = client.configClient.StoreAsset(bosh.CredsFilename, []byte(correctedCreds))
	if err != nil {
		return err
	}
	return nil
}
