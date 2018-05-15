package terraform

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Apply takes a terraform config and applies it
func (client *Client) Apply(dryrun bool) error {
	fmt.Println("RUNNING TERRAFORM APPLY")
	action := "apply"
	if dryrun {
		action = "plan"
	}
	os.Setenv("TF_LOG", "TRACE")
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
	cmd := exec.Command(client.tempDir.Path("terraform"), args...)
	cmd.Dir = client.configDir
	cmd.Stdout = stdout
	cmd.Stderr = client.stderr
	return cmd.Run()
}
