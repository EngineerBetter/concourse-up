package bosh

import (
    "database/sql"
	"fmt"
	"strings"
	//cloud sql proxy package
    _ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
   //metadata for cloud compute
    _ "cloud.google.com/go/compute/metadata"
)

var gcp_db *sql.DB

func (client *Client) createGCPDatabases() error {
    conn := fmt.Sprintf("host=concourse-up:%s:%s user=%s dbname=postgres password=%s sslmode=disable", client.provider.Region(),client.config.RDSDefaultDatabaseName,  client.config.RDSUsername, client.config.RDSPassword)
    gcp_db, err := sql.Open("cloudsqlpostgres", conn)
    if err != nil {
        return err
    }
	defer gcp_db.Close()
	dbNames := []string{"concourse_atc", "uaa", "credhub"}
	for _, dbName := range dbNames {
		_, err := gcp_db.Exec("CREATE DATABASE " + dbName)
		if err != nil && !strings.Contains(err.Error(),
			fmt.Sprintf(`pq: database "%s" already exists`, dbName)) {
			return err
		}
	}
    return nil
}