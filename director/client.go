package director

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/EngineerBetter/concourse-up/util"
)

// IClient represents a bosh director client
type IClient interface {
	RunCommand(stdout, stderr io.Writer, args ...string) error
	RunAuthenticatedCommand(stdout, stderr io.Writer, detach bool, args ...string) error
	SaveFileToWorkingDir(path string, contents []byte) (string, error)
	PathInWorkingDir(filename string) string
	Cleanup() error
}

// Credentials represent credentials for connecting to bosh director
type Credentials struct {
	Username string
	Password string
	Host     string
	CACert   string
}

// Client represents a low-level wrapper for bosh director
type Client struct {
	tempDir             *util.TempDir
	creds               Credentials
	caCertPath          string
	hasDownloadedBinary bool
	versionFile         []byte
}

const caCertFilename = "ca-cert.pem"

// NewClient returns a new client
func NewClient(creds Credentials, versionFile []byte) (*Client, error) {
	tempDir, err := util.NewTempDir()
	if err != nil {
		return nil, err
	}

	caCertPath, err := tempDir.Save(caCertFilename, []byte(creds.CACert))
	if err != nil {
		return nil, err
	}

	return &Client{
		tempDir:             tempDir,
		creds:               creds,
		caCertPath:          caCertPath,
		hasDownloadedBinary: false,
		versionFile:         versionFile,
	}, nil
}

// SaveFileToWorkingDir saves thegiven file to the temporary director working directory
func (client *Client) SaveFileToWorkingDir(filename string, contents []byte) (string, error) {
	return client.tempDir.Save(filename, contents)
}

// PathInWorkingDir returns a path for the file in the directos' working directory
func (client *Client) PathInWorkingDir(filename string) string {
	return client.tempDir.Path(filename)
}

// Cleanup removes tempfiles
func (client *Client) Cleanup() error {
	return client.tempDir.Cleanup()
}

func (client *Client) ensureBinaryDownloaded() error {
	if client.hasDownloadedBinary {
		return nil
	}

	fileHandler, err := os.Create(client.tempDir.Path("bosh-cli"))
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	url, err := getBoshCLIURL(client.versionFile)
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

	client.hasDownloadedBinary = true

	return nil
}

func getBoshCLIURL(versionFile []byte) (string, error) {
	var x map[string]map[string]string
	err := json.Unmarshal(versionFile, &x)
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return x["bosh-cli"]["mac"], nil
	case "linux":
		return x["bosh-cli"]["linux"], nil
	case "windows":
		return x["bosh-cli"]["windows"], nil
	default:
		return "", fmt.Errorf("unknown os: `%s`", runtime.GOOS)
	}
}
