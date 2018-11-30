package bosh


func (client *Client) createDefaultDatabases() error {
	//client.createGCPDatabases()
	switch client.provider.IAAS() {
	case "AWS": //nolint
		return client.createAWSDatabases()
	case "GCP": //nolint
		return client.createGCPDatabases()
	}

return nil

}