package terraform

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/resource"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

// InputVars exposes ConfigureDirectorManifestCPI
type InputVars interface {
	ConfigureTerraform(string) (string, error)
}

// IAASMetadata holds IAAS specific terraform metadata
type IAASMetadata interface {
	AssertValid() error
	Init(*bytes.Buffer) error
	Get(string) (string, error)
}

//CLIInterface is interface the abstraction of execCmd
type CLIInterface interface {
	IAAS(iaas.Name) (InputVars, IAASMetadata, error)
	Apply(InputVars, bool) error
	Destroy(InputVars) error
	BuildOutput(InputVars, IAASMetadata) error
}

// CLI struct holds the abstraction of execCmd
type CLI struct {
	execCmd func(string, ...string) *exec.Cmd
	Path    string
	iaas    iaas.Name
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

// NullInputVars exposes ConfigureDirectorManifestCPI
type NullInputVars struct{}

// ConfigureTerraform is a nullified function
func (n *NullInputVars) ConfigureTerraform(string) (string, error) { return "", nil }

// Build is a nullified function
func (n *NullInputVars) Build(map[string]interface{}) error { return nil }

// NullMetadata holds IAAS specific terraform metadata
type NullMetadata struct{}

// AssertValid is a nullified function
func (n *NullMetadata) AssertValid() error { return nil }

// Init is a nullified function
func (n *NullMetadata) Init(*bytes.Buffer) error { return nil }

// Get is a nullified function
func (n *NullMetadata) Get(string) (string, error) { return "", nil }

// IAAS is returning the IAAS specific metadata and environment
func (c *CLI) IAAS(name iaas.Name) (InputVars, IAASMetadata, error) {
	switch name {
	case iaas.AWS: // nolint
		c.iaas = iaas.AWS // nolint
		return &AWSInputVars{}, &AWSMetadata{}, nil
	case iaas.GCP: // nolint
		c.iaas = iaas.GCP // nolint
		return &GCPInputVars{}, &GCPMetadata{}, nil
	}
	return &NullInputVars{}, &NullMetadata{}, errors.New("terraform: " + name.String() + " not a valid iaas provider")

}

func (c *CLI) init(config InputVars) (string, error) {
	var (
		tfConfig string
		err      error
	)
	switch c.iaas {
	case iaas.AWS: // nolint
		tfConfig, err = config.ConfigureTerraform(resource.AWSTerraformConfig)
		if err != nil {
			return "", err
		}
	case iaas.GCP: // nolint
		tfConfig, err = config.ConfigureTerraform(resource.GCPTerraformConfig)
		if err != nil {
			return "", err
		}
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
