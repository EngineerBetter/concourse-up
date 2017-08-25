package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/util"
)

// DarwinBinaryURL is a compile-time variable set with -ldflags
var DarwinBinaryURL = "COMPILE_TIME_VARIABLE_terraform_darwin_binary_url"

// LinuxBinaryURL is a compile-time variable set with -ldflags
var LinuxBinaryURL = "COMPILE_TIME_VARIABLE_terraform_linux_binary_url"

// WindowsBinaryURL is a compile-time variable set with -ldflags
var WindowsBinaryURL = "COMPILE_TIME_VARIABLE_terraform_windows_binary_url"

// IClient is an interface for the terraform Client
type IClient interface {
	Output() (*Metadata, error)
	Apply(dryrun bool) error
	Destroy() error
	Cleanup() error
}

// Client wraps common terraform commands
type Client struct {
	iaas      string
	configDir string
	tempDir   *util.TempDir
	stdout    io.Writer
	stderr    io.Writer
}

// ClientFactory is a function that builds a client interface
type ClientFactory func(iaas string, config *config.Config, stdout, stderr io.Writer) (IClient, error)

// NewClient is a concrete implementation of ClientFactory
func NewClient(iaas string, config *config.Config, stdout, stderr io.Writer) (IClient, error) {
	if iaas != "AWS" {
		return nil, fmt.Errorf("IAAS not supported: %s", iaas)
	}

	tempDir, err := util.NewTempDir()
	if err != nil {
		return nil, err
	}

	if err := setupBinary(tempDir); err != nil {
		return nil, err
	}

	terraformFile, err := util.RenderTemplate(AWSTemplate, config)
	if err != nil {
		return nil, err
	}

	configDir := tempDir.Path("config")
	if err := os.Mkdir(configDir, 0777); err != nil {
		return nil, err
	}

	configPath := tempDir.Path("config/main.tf")
	if err := ioutil.WriteFile(configPath, terraformFile, 0777); err != nil {
		return nil, err
	}

	client := &Client{
		tempDir:   tempDir,
		configDir: configDir,
		stdout:    stdout,
		stderr:    stderr,
	}
	devNull := bytes.NewBuffer(nil)
	if err := client.terraform([]string{
		"init",
	}, devNull); err != nil {
		// if there is an error dump stdout for debugging
		io.Copy(stdout, devNull)
		return nil, err
	}
	return client, nil
}

// Cleanup cleans up the temporary directory used by terraform
func (client *Client) Cleanup() error {
	return os.RemoveAll(client.configDir)
}

// Output fetches the terraform output/metadata
func (client *Client) Output() (*Metadata, error) {
	stdoutBuffer := bytes.NewBuffer(nil)
	if err := client.terraform([]string{
		"output",
		"-json",
	}, stdoutBuffer); err != nil {
		return nil, err
	}

	metadata := Metadata{}
	if err := json.NewDecoder(stdoutBuffer).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func getTerraformURL() (string, error) {
	os := runtime.GOOS
	if os == "darwin" {
		return DarwinBinaryURL, nil
	} else if os == "linux" {
		return LinuxBinaryURL, nil
	} else if os == "windows" {
		return WindowsBinaryURL, nil
	}
	return "", fmt.Errorf("unknown os: `%s`", os)
}

func setupBinary(tempDir *util.TempDir) error {
	fileHandler, err := os.Create(tempDir.Path("terraform"))
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	url, err := getTerraformURL()
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(fileHandler, resp.Body); err != nil {
		return err
	}

	if err := fileHandler.Sync(); err != nil {
		return err
	}

	if err := os.Chmod(fileHandler.Name(), 0700); err != nil {
		return err
	}

	return nil
}
