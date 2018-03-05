package bosh

import (
	"fmt"

	"strings"
)

func (client *Client) createDefaultDatabases() error {
	dbNames := []string{client.config.ConcourseDBName, "uaa", "credhub"}
	for _, dbName := range dbNames {
		_, err := client.db.Exec("CREATE DATABASE " + dbName)
		if err != nil && !strings.Contains(err.Error(),
			fmt.Sprintf(`pq: database "%s" already exists`, dbName)) {
			return err
		}
	}
	return nil
}
