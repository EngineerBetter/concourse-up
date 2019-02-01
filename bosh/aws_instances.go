package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Instances returns the list of Concourse VMs
func (client *AWSClient) Instances() ([]Instance, error) {
	output := new(bytes.Buffer)

	if err := client.director.RunAuthenticatedCommand(
		output,
		client.stderr,
		false,
		"--deployment",
		concourseDeploymentName,
		"instances",
		"--json",
	); err != nil {
		// if there is an error, copy the stdout to the main stdout to help debugging
		return nil, fmt.Errorf("Error running `bosh instances`: %s", output.String())
	}

	jsonOutput := struct {
		Tables []struct {
			Rows []struct {
				Instance     string `json:"instance"`
				IPs          string `json:"ips"`
				ProcessState string `json:"process_state"`
			} `json:"Rows"`
		} `json:"Tables"`
	}{}

	if err := json.NewDecoder(output).Decode(&jsonOutput); err != nil {
		return nil, err
	}

	instances := []Instance{}

	for _, table := range jsonOutput.Tables {
		for _, row := range table.Rows {
			instances = append(instances, Instance{
				Name:  row.Instance,
				IP:    row.IPs,
				State: row.ProcessState,
			})
		}
	}

	return instances, nil
}
