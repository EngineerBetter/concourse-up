package bosh

import (
	"io/ioutil"
	"os"
)

const pemFilename = "director.pem"
const directorManifestFilename = "director.yml"

// Deploy deploys a new Bosh director or converges an existing deployment
// Returns new contents of bosh state file
func (client *Client) Deploy(state, creds []byte, detach bool) (newState, newCreds []byte, err error) {
	state, creds, err = client.createEnv(state, creds)
	if err != nil {
		return state, creds, err
	}

	if err = client.updateCloudConfig(); err != nil {
		return state, creds, err
	}

	if err = client.uploadConcourseStemcell(); err != nil {
		return state, creds, err
	}

	if err = client.createDefaultDatabases(); err != nil {
		return state, creds, err
	}

	creds, err = client.deployConcourse(creds, detach)
	if err != nil {
		return state, creds, err
	}

	return state, creds, err
}

func touch(name string) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	return f.Close()
}

func (client *Client) createEnv(state, creds []byte) (newState, newCreds []byte, err error) {
	stateFilePath, err := client.saveStateFile(state)
	if err != nil {
		return state, creds, err
	}
	credsFilePath, err := client.saveCredsFile(creds)
	if err != nil {
		return state, creds, err
	}

	_, err = client.director.SaveFileToWorkingDir(pemFilename, []byte(client.config.PrivateKey))
	if err != nil {
		return state, creds, err
	}

	directorManifestBytes, err := generateBoshInitManifest(client.config, client.metadata, pemFilename)
	if err != nil {
		return state, creds, err
	}

	directorManifestPath, err := client.director.SaveFileToWorkingDir(directorManifestFilename, directorManifestBytes)
	if err != nil {
		return state, creds, err
	}

	extraTagsPath, err := client.director.SaveFileToWorkingDir(extraTagsFilename, extraTags)
	if err != nil {
		return state, creds, err
	}

	err = touch(credsFilePath)
	if err != nil {
		return state, creds, err
	}

	flagFiles := []string{
		"create-env",
		directorManifestPath,
		"--state",
		stateFilePath,
		"--vars-store",
		credsFilePath,
	}

	t, err1 := client.buildTagsYaml(client.config.Project, "director")
	if err1 != nil {
		return state, creds, err
	}
	vmap := map[string]interface{}{
		"tags": t,
	}
	vs := vars(vmap)
	flagFiles = append(flagFiles, "--ops-file", extraTagsPath)
	flagFiles = append(flagFiles, vs...)

	if err = client.director.RunCommand(
		client.stdout,
		client.stderr,
		flagFiles...,
	); err != nil {
		// Even if deploy does not work, try and save the state file
		// This prevents issues with re-trying
		state, _ = ioutil.ReadFile(stateFilePath)
		creds, _ = ioutil.ReadFile(credsFilePath)
		return state, creds, err
	}

	state, err = ioutil.ReadFile(stateFilePath)
	if err != nil {
		return state, creds, err
	}
	creds, err = ioutil.ReadFile(credsFilePath)
	return state, creds, err
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

	return client.director.RunAuthenticatedCommand(
		client.stdout,
		client.stderr,
		false,
		"update-cloud-config",
		cloudConfigPath,
	)
}

func (client *Client) saveStateFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(StateFilename), nil
	}

	return client.director.SaveFileToWorkingDir(StateFilename, bytes)
}

func (client *Client) saveCredsFile(bytes []byte) (string, error) {
	if bytes == nil {
		return client.director.PathInWorkingDir(CredsFilename), nil
	}

	return client.director.SaveFileToWorkingDir(CredsFilename, bytes)
}
