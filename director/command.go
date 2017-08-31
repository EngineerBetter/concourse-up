package director

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

var defaultBoshArgs = []string{"--non-interactive", "--tty", "--no-color"}

// RunAuthenticatedCommand runs a command against the bosh director, after authenticating
func (client *Client) RunAuthenticatedCommand(stdout, stderr io.Writer, detach bool, args ...string) error {
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

	if detach {
		return client.runDetachingCommand(stdout, stderr, args...)
	}

	return client.RunCommand(stdout, stderr, args...)
}

// RunCommand runs a command against the bosh director
// https://github.com/cloudfoundry/bosh-cli/blob/master/main.go
func (client *Client) RunCommand(stdout, stderr io.Writer, args ...string) error {
	if err := client.ensureBinaryDownloaded(); err != nil {
		return err
	}

	args = append(defaultBoshArgs, args...)

	cmd := exec.Command(client.tempDir.Path("bosh-cli"), args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (client *Client) runDetachingCommand(stdout, stderr io.Writer, args ...string) error {
	if err := client.ensureBinaryDownloaded(); err != nil {
		return err
	}

	args = append(defaultBoshArgs, args...)

	cmd := exec.Command(client.tempDir.Path("bosh-cli"), args...)
	cmd.Stderr = stderr

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)

	if err := cmd.Start(); err != nil {
		return err
	}

	for scanner.Scan() {
		text := scanner.Text()
		if err := stdout.Write([]byte(fmt.Sprintf("%s\n", text))); err != nil {
			return err
		}
		if strings.HasPrefix(text, "Task") {
			stdout.Write([]byte("Task started, detaching output\n"))
			return nil
		}
	}

	return fmt.Errorf("Didn't detect successful task start in BOSH comand: bosh-cli %s", strings.Join(args, " "))
}
