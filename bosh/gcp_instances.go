package bosh

// Instances returns the list of Concourse VMs
func (client *GCPClient) Instances() ([]Instance, error) {
	return Instances(client, client.director, client.stderr)
}
