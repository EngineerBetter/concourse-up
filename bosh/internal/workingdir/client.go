package workingdir

import (
	"github.com/EngineerBetter/concourse-up/util"
)

//go:generate counterfeiter . IClient
// IClient represents client for interacting with a working directory
type IClient interface {
	SaveFileToWorkingDir(path string, contents []byte) (string, error)
	PathInWorkingDir(filename string) string
	Cleanup() error
}

// client represents a low-level wrapper for a working directory
type client struct {
	tempDir *util.TempDir
}

// New returns a new client
func New() (*client, error) {
	tempDir, err := util.NewTempDir()
	if err != nil {
		return nil, err
	}

	return &client{
		tempDir: tempDir,
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
