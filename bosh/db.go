package bosh

import (
	"fmt"

	"strings"
)

func (client *Client) createDefaultDatabases() error {
	dbNames := []string{client.config.ConcourseDBName, "uaa", "credhub"}

	for _, dbName := range dbNames {
		err := client.dbRunner(fmt.Sprintf("CREATE DATABASE %s;", dbName))

		if err != nil && !strings.Contains(err.Error(),
			fmt.Sprintf("pq: database \"%s\" already exists", dbName)) {
			return err
		}
	}
	return nil
}
