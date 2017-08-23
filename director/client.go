package director

import (
	"io"

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
	tempDir    *util.TempDir
	creds      Credentials
	caCertPath string
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
		tempDir:    tempDir,
		creds:      creds,
		caCertPath: caCertPath,
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
