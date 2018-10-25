package terraform

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/EngineerBetter/concourse-up/terraform/internal/aws"

	"github.com/EngineerBetter/concourse-up/resource"
)

// InputVars exposes ConfigureDirectorManifestCPI
type InputVars interface {
	ConfigureTerraform(string) (string, error)
	Build(map[string]interface{}) error
}

// IAASMetadata holds IAAS specific terraform metadata
type IAASMetadata interface {
	AssertValid() error
	Init(*bytes.Buffer) error
	Get(string) (string, error)
}

//CLIInterface is interface the abstraction of execCmd
type CLIInterface interface {
	IAAS(string) (InputVars, IAASMetadata)
	Apply(InputVars, bool) error
	Destroy(InputVars) error
	BuildOutput(InputVars, IAASMetadata) error
}

// CLI struct holds the abstraction of execCmd
type CLI struct {
	execCmd func(string, ...string) *exec.Cmd
	Path    string
}

// Option defines the arbitary element of Options for New
type Option func(*CLI) error

// Path returns the path of the terraform-cli as an Option
func Path(path string) Option {
	return func(c *CLI) error {
		c.Path = path
		return nil
	}
}

// DownloadTerraform returns the dowloaded CLI path Option
func DownloadTerraform() Option {
	return func(c *CLI) error {
		path, err := resource.TerraformCLIPath()
		c.Path = path
		return err
	}
}

// New provides a new CLI
func New(ops ...Option) (*CLI, error) {
	// @Note: we will have to switch between IAASs at this point
	// for the time being we are using directly AWS
	cli := &CLI{
		execCmd: exec.Command,
		Path:    "terraform",
	}
	for _, op := range ops {
		if err := op(cli); err != nil {
			return nil, err
		}
	}
	return cli, nil
}

// IAAS is returning the IAAS specific metadata and environment
func (c *CLI) IAAS(name string) (InputVars, IAASMetadata) {
	return &aws.InputVars{}, &aws.Metadata{}
}

func (c *CLI) init(config InputVars) (string, error) {

	tfConfig, err := config.ConfigureTerraform(resource.AWSTerraformConfig)
	if err != nil {
		return "", err
	}

	terraformConfigPath, err := writeTempFile([]byte(tfConfig))
	if err != nil {
		return "", err
	}
	cmd := c.execCmd(c.Path, "init")
	cmd.Dir = terraformConfigPath
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		os.RemoveAll(terraformConfigPath)
		return "", err
	}
	return terraformConfigPath, nil
}

// Apply runs terraform apply for a given config
func (c *CLI) Apply(config InputVars, dryrun bool) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	defer os.RemoveAll(terraformConfigPath)

	action := "apply"
	if dryrun {
		action = "plan"
	}

	cmd := c.execCmd(c.Path, action, "-input=false", "-auto-approve")
	cmd.Dir = terraformConfigPath

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

// Destroy destroys terraform resources specified in a config file
func (c *CLI) Destroy(config InputVars) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	defer os.RemoveAll(terraformConfigPath)

	cmd := c.execCmd(c.Path, "destroy", "-auto-approve")
	cmd.Dir = terraformConfigPath
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// BuildOutput builds the terraform output/metadata
func (c *CLI) BuildOutput(config InputVars, metadata IAASMetadata) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	defer os.RemoveAll(terraformConfigPath)

	stdoutBuffer := bytes.NewBuffer(nil)
	cmd := c.execCmd(c.Path, "output", "-json")
	cmd.Dir = terraformConfigPath
	cmd.Stderr = os.Stderr
	cmd.Stdout = stdoutBuffer
	if err := cmd.Run(); err != nil {
		return err
	}

	return metadata.Init(stdoutBuffer)
}

func writeTempFile(data []byte) (string, error) {
	mode := int(0740)
	perm := os.FileMode(mode)
	dirName := randomString()
	filePath := path.Join(os.TempDir(), dirName)
	err := os.MkdirAll(filePath, perm)
	if err != nil {
		return "", err
	}
	f, err := ioutil.TempFile(filePath, "*.tf")
	if err != nil {
		return "", err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.RemoveAll(filePath)
		return "", err
	}
	return filePath, err
}

func randomString() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x%x%x%x", b[0:2], b[2:4], b[4:6], b[6:8])
}
