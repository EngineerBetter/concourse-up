package bosh

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

// Delete deletes a bosh director
func (client *Client) Delete(stateFileBytes []byte) ([]byte, error) {
	if err := client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"--deployment",
		concourseDeploymentName,
		"delete-deployment",
		"--force",
	); err != nil {
		return nil, err
	}

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

	output := bytes.NewBuffer(nil)
	stdout := io.MultiWriter(client.stdout, output)
	if err := client.director.RunCommand(
		stdout,
		client.stderr,
		"delete-env",
		directorManifestPath,
		"--state",
		stateFilePath,
	); err != nil {
		return stateFileBytes, err
	}

	if !strings.Contains(output.String(), "Finished deleting deployment") {
		return nil, errors.New("Couldn't find string `Finished deleting deployment` in bosh stdout/stderr output")
	}

	return nil, nil
}
