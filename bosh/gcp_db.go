package bosh

func (client *GCPClient) createDefaultDatabases() error {
	return client.provider.CreateDatabases(client.config.RDSDefaultDatabaseName, client.config.RDSUsername, client.config.RDSPassword)
}
