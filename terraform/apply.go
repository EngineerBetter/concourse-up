package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const appName string = "concourse-up"

// IClient is an interface for the terraform Client
type IClient interface {
	Output() (*Metadata, error)
	Apply() error
	Destroy() error
	Cleanup() error
}

// Client wraps common terraform commands
type Client struct {
	iaas      string
	configDir string
	stdout    io.Writer
	stderr    io.Writer
}

// ClientFactory is a function that builds a client interface
type ClientFactory func(iaas string, config []byte, stdout, stderr io.Writer) (IClient, error)

// NewClient is a concrete implementation of ClientFactory
func NewClient(iaas string, config []byte, stdout, stderr io.Writer) (IClient, error) {
	if err := checkTerraformOnPath(stderr, stderr); err != nil {
		return nil, err
	}

	configDir, err := initConfig(config, stderr, stderr)
	if err != nil {
		return nil, err
	}

	return &Client{
		iaas:      iaas,
		configDir: configDir,
		stdout:    stdout,
		stderr:    stderr,
	}, nil
}

// Cleanup cleans up the temporary directory used by terraform
func (client *Client) Cleanup() error {
	return os.RemoveAll(client.configDir)
}

// Output fetches the terraform output/metadata
func (client *Client) Output() (*Metadata, error) {
	stdoutBuffer := bytes.NewBuffer(nil)
	if err := terraform([]string{
		"output",
		"-json",
	}, client.configDir, stdoutBuffer, client.stderr); err != nil {
		return nil, err
	}

	metadata := Metadata{
		AWS: &AWSMetadata{},
	}
	if err := json.NewDecoder(stdoutBuffer).Decode(metadata.AWS); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// Apply takes a terraform config and applies it
func (client *Client) Apply() error {
	return terraform([]string{
		"apply",
		"-input=false",
	}, client.configDir, client.stdout, client.stderr)
}

// Destroy destroys the given terraform config
func (client *Client) Destroy() error {
	return terraform([]string{
		"destroy",
		"-force",
	}, client.configDir, client.stdout, client.stderr)
}

func checkTerraformOnPath(stdout, stderr io.Writer) error {
	if err := terraform([]string{"version"}, "", stdout, stderr); err != nil {
		return fmt.Errorf("error running `terraform version`, is terraform in your PATH? Download terraform here: https://www.terraform.io/downloads.html\n%s", err.Error())
	}
	return nil
}

func terraform(args []string, dir string, stdout, stderr io.Writer) error {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func initConfig(config []byte, stdout, stderr io.Writer) (string, error) {
	// write out config
	tmpDir, err := ioutil.TempDir("", appName)
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(tmpDir, "main.tf")
	if err := ioutil.WriteFile(configPath, config, 0777); err != nil {
		return "", err
	}

	if err := terraform([]string{
		"init",
	}, tmpDir, stdout, stderr); err != nil {
		return "", err
	}

	return tmpDir, nil
}
