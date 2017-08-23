package director

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/EngineerBetter/concourse-up/util"
)

// DarwinBinaryURL is a compile-time variable set with -ldflags
var DarwinBinaryURL = "COMPILE_TIME_VARIABLE_director_darwin_binary_url"

// LinuxBinaryURL is a compile-time variable set with -ldflags
var LinuxBinaryURL = "COMPILE_TIME_VARIABLE_director_linux_binary_url"

// WindowsBinaryURL is a compile-time variable set with -ldflags
var WindowsBinaryURL = "COMPILE_TIME_VARIABLE_director_windows_binary_url"

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
}

const caCertFilename = "ca-cert.pem"

// NewClient returns a new client
func NewClient(creds Credentials) (*Client, error) {
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

	url, err := getBoshCLIURL()
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

func getBoshCLIURL() (string, error) {
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
