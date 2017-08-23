package director

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
)

var defaultBoshArgs = []string{"--non-interactive", "--tty", "--no-color"}

// RunAuthenticatedCommand runs a command against the bosh director, after authenticating
func (client *Client) RunAuthenticatedCommand(stdout, stderr io.Writer, detach bool, args ...string) error {
	if detach {
		return errors.New("detach mode not yet implemented")
	}
	if err := client.ensureBinaryDownloaded(); err != nil {
		return err
	}
	args = append([]string{
		"--environment",
		fmt.Sprintf("https://%s", client.creds.Host),
		"--ca-cert",
		client.caCertPath,
		"--client",
		client.creds.Username,
		"--client-secret",
		client.creds.Password,
	}, args...)

	return client.RunCommand(stdout, stderr, args...)
}

// RunCommand runs a command against the bosh director
// https://github.com/cloudfoundry/bosh-cli/blob/master/main.go
func (client *Client) RunCommand(stdout, stderr io.Writer, args ...string) error {
	if err := client.ensureBinaryDownloaded(); err != nil {
		return err
	}

	cmd := exec.Command(client.tempDir.Path("bosh-cli"), args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
