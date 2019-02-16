package bosh

import "fmt"

// Instances returns the list of Concourse VMs
func (client *GCPClient) Instances() ([]Instance, error) {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve director IP: [%v]", err)
	}

	return instances(
		client.boshCLI,
		directorPublicIP,
		client.config.DirectorPassword,
		client.config.DirectorCACert,
	)
}
