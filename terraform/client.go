package terraform

import (
	"archive/zip"
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

//go:generate go-bindata -pkg $GOPACKAGE assets/ ../resources/director-versions.json

// AWSTemplate is a terraform configuration template for AWS
var AWSTemplate = string(MustAsset("assets/main.tf"))

// IClient is an interface for the terraform Client
type IClient interface {
	Output() (*Metadata, error)
	Apply(dryrun bool) error
	Destroy() error
	Cleanup() error
}

// Client wraps common terraform commands
type Client struct {
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

var versionFile = MustAsset("../resources/director-versions.json")

func getTerraformURL() (string, error) {
	var x map[string]map[string]string
	err := json.Unmarshal(versionFile, &x)
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return x["terraform"]["mac"], nil
	case "linux":
		return x["terraform"]["linux"], nil
	case "windows":
		return x["terraform"]["windows"], nil
	default:
		return "", fmt.Errorf("unknown os: `%s`", runtime.GOOS)
	}
}

func setupBinary(tempDir *util.TempDir) error {
	url, err := getTerraformURL()
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	n, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), n)
	if err != nil {
		return err
	}
	r, err := zr.File[0].Open()
	if err != nil {
		return err
	}
	defer r.Close()

	f, err := os.OpenFile(tempDir.Path("terraform"), os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return err
	}

	return f.Close()
}
