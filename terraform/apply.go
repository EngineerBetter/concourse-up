package terraform

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

// Apply takes a terraform config and applies it
func (client *Client) Apply(dryrun bool) error {
	action := "apply"
	if dryrun {
		action = "plan"
	}
	return terraform([]string{
		action,
		"-input=false",
	}, client.configDir, client.stdout, client.stderr)
}

// Destroy destroys the given terraform config
func (client *Client) Destroy() error {
	return terraform([]string{
		"destroy",
		"-force",
	}, client.configDir, client.stdout, client.stderr)
}

func checkTerraformOnPath(stdout, stderr io.Writer) error {
	if err := terraform([]string{"version"}, "", stdout, stderr); err != nil {
		return fmt.Errorf("error running `terraform version`, is terraform in your PATH? Download terraform here: https://www.terraform.io/downloads.html\n%s", err.Error())
	}
	return nil
}

func terraform(args []string, dir string, stdout, stderr io.Writer) error {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func initConfig(config []byte, stdout, stderr io.Writer) (string, error) {
	// write out config
	tmpDir, err := ioutil.TempDir("", appName)
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(tmpDir, "main.tf")
	if err := ioutil.WriteFile(configPath, config, 0777); err != nil {
		return "", err
	}

	if err := terraform([]string{
		"init",
	}, tmpDir, stdout, stderr); err != nil {
		return "", err
	}

	return tmpDir, nil
}
