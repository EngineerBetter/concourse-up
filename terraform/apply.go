package terraform

import (
	"fmt"
	"io"
	"os/exec"
)

// Apply takes a terraform config and applies it
func (client *Client) Apply(dryrun bool) error {
	action := "apply"
	if dryrun {
		action = "plan"
	}
	return client.terraform([]string{
		action,
		"-input=false",
	}, client.stdout)
}

// Destroy destroys the given terraform config
func (client *Client) Destroy() error {
	return client.terraform([]string{
		"destroy",
		"-force",
	}, client.stdout)
}

func (client *Client) terraform(args []string, stdout io.Writer) error {
	client.stderr.Write([]byte(fmt.Sprintf("running terraform with: %s\n", args)))
	cmd := exec.Command(client.tempDir.Path("terraform"), args...)
	cmd.Dir = client.configDir
	cmd.Stdout = stdout
	cmd.Stderr = client.stderr
	return cmd.Run()
}
