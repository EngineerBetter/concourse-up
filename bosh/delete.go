package bosh

import (
	"errors"
	"io/ioutil"
	"strings"
)

// Delete deletes a bosh director
func (client *Client) Delete(stateFileBytes []byte) ([]byte, error) {
	stateFilePath, err := client.saveStateFile(stateFileBytes)
	if err != nil {
		return stateFileBytes, err
	}

	_, err = client.director.SaveFileToWorkingDir(pemFilename, []byte(client.config.PrivateKey))
	if err != nil {
		return stateFileBytes, err
	}
	directorManifestBytes, err := generateBoshInitManifest(client.config, client.metadata, pemFilename)
	if err != nil {
		return stateFileBytes, err
	}

	directorManifestPath, err := client.director.SaveFileToWorkingDir(directorManifestFilename, directorManifestBytes)
	if err != nil {
		return stateFileBytes, err
	}

	output, err := client.director.RunCommand(
		"delete-env",
		directorManifestPath,
		"--state",
		stateFilePath,
	)
	if err != nil {
		return stateFileBytes, err
	}

	stateFileBytes, err = ioutil.ReadFile(stateFilePath)
	if err != nil {
		return stateFileBytes, err
	}

	if !strings.Contains(string(output), "Finished deleting deployment") {
		return stateFileBytes, errors.New("Couldn't find string `Finished deleting deployment` in bosh stdout/stderr output")
	}

	return stateFileBytes, nil
}
