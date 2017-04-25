package bosh

import (
	"errors"
	"io/ioutil"
	"strings"
)

const pemFilename = "director.pem"
const directorManifestFilename = "director.yml"

// Deploy deploys a new Bosh director or converges an existing deployment
// Returns new contents of bosh state file
func (client *Client) Deploy(stateFileBytes []byte) ([]byte, error) {
	stateFileBytes, err := client.createEnv(stateFileBytes)
	if err != nil {
		return stateFileBytes, err
	}

	if err := client.updateCloudConfig(); err != nil {
		return stateFileBytes, err
	}

	if err := client.uploadConcourse(); err != nil {
		return stateFileBytes, err
	}

	if err := client.createDefaultDatabases(); err != nil {
		return stateFileBytes, err
	}

	if err := client.deployConcourse(); err != nil {
		return stateFileBytes, err
	}

	return stateFileBytes, nil
}

func (client *Client) createEnv(stateFileBytes []byte) ([]byte, error) {
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
		"create-env",
		directorManifestPath,
		"--state",
		stateFilePath,
	)
	if err != nil {
		// Even if deploy does not work, try and save the state file
		// This prevents issues with re-trying
		stateFileBytes, _ = ioutil.ReadFile(stateFilePath)
		return stateFileBytes, err
	}

	stateFileBytes, err = ioutil.ReadFile(stateFilePath)
	if err != nil {
		return stateFileBytes, err
	}

	if !strings.Contains(string(output), "Finished deploying") && !strings.Contains(string(output), "Skipping deploy") {
		return stateFileBytes, errors.New("Couldn't find string `Finished deploying` or `Skipping deploy` in bosh stdout/stderr output")
	}

	return stateFileBytes, nil
}

func (client *Client) updateCloudConfig() error {
	cloudConfigBytes, err := generateCloudConfig(client.config, client.metadata)
	if err != nil {
		return err
	}

	cloudConfigPath, err := client.director.SaveFileToWorkingDir(cloudConfigFilename, cloudConfigBytes)
	if err != nil {
		return err
	}

	_, err = client.director.RunAuthenticatedCommand(
		"update-cloud-config",
		cloudConfigPath,
	)

	return err
}

func (client *Client) saveStateFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(StateFilename), nil
	}

	return client.director.SaveFileToWorkingDir(StateFilename, bytes)
}
