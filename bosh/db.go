package bosh

import (
	"fmt"

	"strings"
)

func (client *Client) createDefaultDatabases() error {
	dbNames := []string{client.config.ConcourseDBName, "uaa", "credhub"}
	stmt, err := client.db.Prepare("CREATE DATABASE $1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, dbName := range dbNames {
		_, err := stmt.Exec(dbName)
		if err != nil && !strings.Contains(err.Error(),
			fmt.Sprintf(`pq: database "%s" already exists`, dbName)) {
			return err
		}
	}
	return nil
}
