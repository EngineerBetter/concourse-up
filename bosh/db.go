package bosh

func (client *Client) createDefaultDatabases() error {
	switch client.provider.IAAS() {
	case "AWS": //nolint
		return client.createAWSDatabases()
	case "GCP": //nolint
		client.provider.CreateDatabases(client.config.RDSDefaultDatabaseName, client.config.RDSUsername, client.config.RDSPassword)
		return nil
	}

	return nil

}
