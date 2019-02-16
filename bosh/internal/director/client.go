package director

import (
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
