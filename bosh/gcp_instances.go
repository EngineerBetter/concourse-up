package bosh

// Instances returns the list of Concourse VMs
func (client *GCPClient) Instances() ([]Instance, error) {
	return instances(client.director, client.stderr)
}
