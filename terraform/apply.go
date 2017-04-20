package terraform

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const appName string = "concourse-up"

// Applier is a function that applies terraform configurations (can be either create or destroy)
type Applier func(config []byte, stdout, stderr io.Writer) error

// Apply takes a terraform config and applies it
func Apply(config []byte, stdout, stderr io.Writer) error {
	if err := checkTerraformOnPath(stdout, stderr); err != nil {
		return err
	}

	configDir, err := initConfig(config, stdout, stderr)
	if err != nil {
		return nil
	}

	if err := terraform([]string{
		"apply",
		"-input=false",
	}, configDir, stdout, stderr); err != nil {
		return err
	}

	if err := cleanup(configDir); err != nil {
		return err
	}

	return nil
}

// Destroy destroys the given terraform config
func Destroy(config []byte, stdout, stderr io.Writer) error {
	if err := checkTerraformOnPath(stdout, stderr); err != nil {
		return err
	}

	configDir, err := initConfig(config, stdout, stderr)
	if err != nil {
		return nil
	}

	// run terraform
	if err := terraform([]string{
		"destroy",
		"-force",
	}, configDir, stdout, stderr); err != nil {
		return err
	}

	if err := cleanup(configDir); err != nil {
		return err
	}

	return nil
}

func checkTerraformOnPath(stdout, stderr io.Writer) error {
	if err := terraform([]string{"version"}, "", stdout, stderr); err != nil {
		return fmt.Errorf("Error running `terraform version`, is terraform in your PATH?\n%s", err.Error())
	}
	return nil
}

func terraform(args []string, dir string, stdout, stderr io.Writer) error {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
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

func cleanup(configDir string) error {
	return os.RemoveAll(configDir)
}
