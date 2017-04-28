package bosh

import "encoding/json"

type Instance struct {
	Name  string
	IP    string
	State string
}

func (client *Client) Instances() ([]Instance, error) {
	output, err := client.director.RunAuthenticatedCommand(
		"--deployment",
		concourseDeploymentName,
		"instances",
		"--json",
	)
	if err != nil {
		return nil, err
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

	if err = json.Unmarshal(output, &jsonOutput); err != nil {
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

	return instances, err
}
