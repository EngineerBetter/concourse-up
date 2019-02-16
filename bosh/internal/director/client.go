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

//go:generate counterfeiter . IClient
// IClient represents a bosh director client
type IClient interface {
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

// client represents a low-level wrapper for bosh director
type client struct {
	tempDir             *util.TempDir
	creds               Credentials
	caCertPath          string
	hasDownloadedBinary bool
	versionFile         []byte
}

const caCertFilename = "ca-cert.pem"

// New returns a new client
func New(creds Credentials, versionFile []byte) (*client, error) {
	tempDir, err := util.NewTempDir()
	if err != nil {
		return nil, err
	}

	caCertPath, err := tempDir.Save(caCertFilename, []byte(creds.CACert))
	if err != nil {
		return nil, err
	}

	return &client{
		tempDir:             tempDir,
		creds:               creds,
		caCertPath:          caCertPath,
		hasDownloadedBinary: false,
		versionFile:         versionFile,
	}, nil
}

// SaveFileToWorkingDir saves the given file to the temporary director working directory
func (client *client) SaveFileToWorkingDir(filename string, contents []byte) (string, error) {
	return client.tempDir.Save(filename, contents)
}

// PathInWorkingDir returns a path for the file in the directos' working directory
func (client *client) PathInWorkingDir(filename string) string {
	return client.tempDir.Path(filename)
}

// Cleanup removes tempfiles
func (client *client) Cleanup() error {
	return client.tempDir.Cleanup()
}

func (client *client) ensureBinaryDownloaded() error {
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
	default:
		return "", fmt.Errorf("unknown os: `%s`", runtime.GOOS)
	}
}
