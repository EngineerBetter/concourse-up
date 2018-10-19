package terraform

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/EngineerBetter/concourse-up/terraform/internal/aws"

	"github.com/EngineerBetter/concourse-up/resource"
)

// TerraformInputVars exposes ConfigureDirectorManifestCPI
type TerraformInputVars interface {
	ConfigureTerraform(string) (string, error)
	Build(map[string]interface{}) error
}

// IAASMetadata holds IAAS specific terraform metadata
type IAASMetadata interface {
	AssertValid() error
	Init(*bytes.Buffer) error
	Get(string) (string, error)
}

// TerraformCLI struct holds the abstraction of execCmd
type TerraformCLI struct {
	execCmd       func(string, ...string) *exec.Cmd
	terraformPath string
}

// Option defines the arbitary element of Options for New
type Option func(*TerraformCLI) error

// TerraformPath returns the path of the terraform-cli as an Option
func TerraformPath(path string) Option {
	return func(c *TerraformCLI) error {
		c.terraformPath = path
		return nil
	}
}

// DownloadTerraform returns the dowloaded TerraformCLI path Option
func DownloadTerraform() Option {
	return func(c *TerraformCLI) error {
		path, err := resource.TerraformCLIPath()
		c.terraformPath = path
		return err
	}
}

// New provides a new TerraformCLI
func New(ops ...Option) (*TerraformCLI, error) {
	// @Note: we will have to switch between IAASs at this point
	// for the time being we are using directly AWS
	c := &TerraformCLI{
		execCmd:       exec.Command,
		terraformPath: "terraform",
	}
	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// IAAS is returning the IAAS specific metadata and environment
func (c *TerraformCLI) IAAS(name string) (TerraformInputVars, IAASMetadata) {
	return &aws.InputVars{}, &aws.Metadata{}
}

// Store exposes its methods
type Store interface {
	Set(key string, value []byte) error
	// Get must return a zero length byte slice and a nil error when the key is not present in the store
	Get(string) ([]byte, error)
}

func (c *TerraformCLI) init(config TerraformInputVars) (string, error) {

	tfConfig, err := config.ConfigureTerraform(resource.AWSTerraformConfig)
	if err != nil {
		return "", err
	}
	terraformConfigPath, err := writeTempFile([]byte(tfConfig))
	if err != nil {
		return "", err
	}
	terraformConfigDir := path.Dir(terraformConfigPath)

	cmd := c.execCmd(c.terraformPath, "init")
	cmd.Dir = terraformConfigDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		os.Remove(terraformConfigPath)
		return "", err
	}
	return terraformConfigPath, nil
}

// Apply runs terraform apply for a given config
func (c *TerraformCLI) Apply(config TerraformInputVars, dryrun bool) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	terraformConfigDir := path.Dir(terraformConfigPath)
	defer os.Remove(terraformConfigPath)

	action := "apply"
	if dryrun {
		action = "plan"
	}

	cmd := c.execCmd(c.terraformPath, action, "-input=false", "-auto-approve")
	cmd.Dir = terraformConfigDir

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

// Destroy destroys terraform resources specified in a config file
func (c *TerraformCLI) Destroy(config TerraformInputVars) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	terraformConfigDir := path.Dir(terraformConfigPath)
	defer os.Remove(terraformConfigPath)

	cmd := c.execCmd(c.terraformPath, "destroy", "-auto-approve")
	cmd.Dir = terraformConfigDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// BuildOutput builds the terraform output/metadata
func (c *TerraformCLI) BuildOutput(config TerraformInputVars, metadata IAASMetadata) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	terraformConfigDir := path.Dir(terraformConfigPath)
	defer os.Remove(terraformConfigPath)

	stdoutBuffer := bytes.NewBuffer(nil)
	cmd := c.execCmd(c.terraformPath, "output", "-json")
	cmd.Dir = terraformConfigDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = stdoutBuffer
	if err := cmd.Run(); err != nil {
		return err
	}

	return metadata.Init(stdoutBuffer)
}

func writeTempFile(data []byte) (string, error) {
	f, err := ioutil.TempFile("", "*.tf")
	if err != nil {
		return "", err
	}
	name := f.Name()
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(name)
	}
	return name, err
}
