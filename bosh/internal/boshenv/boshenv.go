package boshenv

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/EngineerBetter/concourse-up/bosh/internal/resource"
	"github.com/EngineerBetter/concourse-up/bosh/internal/yaml"
)

// BOSHCLI struct holds the abstraction of execCmd
type BOSHCLI struct {
	execCmd  func(string, ...string) *exec.Cmd
	boshPath string
}

// Option defines the arbitary element of Options for New
type Option func(*BOSHCLI) error

// BOSHPath returns the path of the bosh-cli as an Option
func BOSHPath(path string) Option {
	return func(c *BOSHCLI) error {
		c.boshPath = path
		return nil
	}
}

// DownloadBOSH returns the dowloaded boshcli path Option
func DownloadBOSH() Option {
	return func(c *BOSHCLI) error {
		path, err := resource.BOSHCLIPath()
		c.boshPath = path
		return err
	}
}

// New provides a new BOSHCLI
func New(ops ...Option) (*BOSHCLI, error) {
	c := &BOSHCLI{
		execCmd:  exec.Command,
		boshPath: "bosh",
	}
	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// IAASEnvironment exposes ConfigureDirectorManifestCPI
type IAASEnvironment interface {
	ConfigureDirectorManifestCPI(string) (string, error)
	ConfigureDirectorCloudConfig(string) (string, error)
}

// Store exposes its methods
type Store interface {
	Set(key string, value []byte) error
	// Get must return a zero length byte slice and a nil error when the key is not present in the store
	Get(string) ([]byte, error)
}

func (c *BOSHCLI) xEnv(action string, store Store, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) error {
	const stateFilename = "state.json"
	const varsFilename = "vars.yaml"
	manifest, err := config.ConfigureDirectorManifestCPI(resource.DirectorManifest)
	if err != nil {
		return err
	}
	boshResource := resource.Get(resource.BOSHRelease)
	bpmResource := resource.Get(resource.BPMRelease)

	vars := map[string]interface{}{
		"director_name":            "bosh",
		"admin_password":           password,
		"director_ssl.certificate": cert,
		"director_ssl.private_key": key,
		"director_ssl.ca":          ca,
		"bosh_url":                 boshResource.URL,
		"bosh_version":             boshResource.Version,
		"bosh_sha1":                boshResource.SHA1,
		"bpm_url":                  bpmResource.URL,
		"bpm_version":              bpmResource.Version,
		"bpm_sha1":                 bpmResource.SHA1,
		"tags":                     tags,
	}
	manifest, err = yaml.Interpolate(manifest, "", vars)
	if err != nil {
		return err
	}
	statePath, uploadState, err := writeToDisk(store, stateFilename)
	if err != nil {
		return err
	}
	defer uploadState()
	varsPath, uploadVars, err := writeToDisk(store, varsFilename)
	if err != nil {
		return err
	}
	defer uploadVars()
	manifestPath, err := writeTempFile([]byte(manifest))
	if err != nil {
		return err
	}
	defer os.Remove(manifestPath)

	cmd := c.execCmd(c.boshPath, action, "--state="+statePath, "--vars-store="+varsPath, manifestPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (c *BOSHCLI) UpdateCloudConfig(config IAASEnvironment, password, cert, key, ca string) error {
	cloudConfig, err := config.ConfigureDirectorCloudConfig(resource.AWSDirectorCloudConfig)
	if err != nil {
		return err
	}

	cloudConfigPath, err := writeTempFile([]byte(cloudConfig))
	if err != nil {
		return err
	}
	defer os.Remove(cloudConfigPath)

	cmd := c.execCmd(c.boshPath, "update-cloud-config", cloudConfigPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// DeleteEnv deletes a bosh env
func (c *BOSHCLI) DeleteEnv(store Store, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) error {
	return c.xEnv("delete-env", store, config, password, cert, key, ca, tags)
}

// CreateEnv creates a bosh env
func (c *BOSHCLI) CreateEnv(store Store, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) error {
	return c.xEnv("create-env", store, config, password, cert, key, ca, tags)

}

func writeToDisk(store Store, key string) (filename string, upload func() error, err error) {
	data, err := store.Get(key)
	if err != nil {
		return "", nil, err
	}
	var path string
	if len(data) == 0 {
		path, err = ioutil.TempDir("", "")
		path = filepath.Join(path, key)
	} else {
		path, err = writeTempFile(data)
	}
	if err != nil {
		return "", nil, err
	}
	upload = func() error {
		defer os.Remove(path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		return store.Set(key, data)
	}
	return path, upload, nil
}

func writeTempFile(data []byte) (string, error) {
	f, err := ioutil.TempFile("", "")
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
