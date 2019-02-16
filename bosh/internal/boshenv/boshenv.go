package boshenv

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/EngineerBetter/concourse-up/resource"
	"github.com/EngineerBetter/concourse-up/util/yaml"
)

const awsConst = "AWS"

const gcpConst = "GCP"

//go:generate counterfeiter . IBOSHCLI
type IBOSHCLI interface {
	CreateEnv(store Store, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) error
	DeleteEnv(store Store, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) error
	RunAuthenticatedCommand(action, ip, password, ca string, detach bool, stdout io.Writer, flags ...string) error
	Locks(config IAASEnvironment, ip, password, ca string) ([]byte, error)
	Recreate(config IAASEnvironment, ip, password, ca string) error
	UpdateCloudConfig(config IAASEnvironment, ip, password, ca string) error
	UploadConcourseStemcell(config IAASEnvironment, ip, password, ca string) error
}

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
func New(ops ...Option) (IBOSHCLI, error) {
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
	ConfigureConcourseStemcell(string) (string, error)
	IAASCheck() string
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

// UpdateCloudConfig generates cloud config from template and use it to update bosh cloud config
func (c *BOSHCLI) UpdateCloudConfig(config IAASEnvironment, ip, password, ca string) error {
	var cloudConfig string
	var err error
	iaas := config.IAASCheck()
	switch iaas {
	case awsConst:
		cloudConfig, err = config.ConfigureDirectorCloudConfig(resource.AWSDirectorCloudConfig)

	case gcpConst:
		cloudConfig, err = config.ConfigureDirectorCloudConfig(resource.GCPDirectorCloudConfig)
	}
	if err != nil {
		return err
	}
	cloudConfigPath, err := writeTempFile([]byte(cloudConfig))
	if err != nil {
		return err
	}
	defer os.Remove(cloudConfigPath)
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)
	cmd := c.execCmd(c.boshPath, "--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "update-cloud-config", cloudConfigPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// Locks runs bosh locks
func (c *BOSHCLI) Locks(config IAASEnvironment, ip, password, ca string) ([]byte, error) {
	var out bytes.Buffer
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return nil, err
	}
	defer os.Remove(caPath)
	cmd := c.execCmd(c.boshPath, "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "locks", "--json")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// UploadConcourseStemcell uploads a stemcell for the chosen IAAS
func (c *BOSHCLI) UploadConcourseStemcell(config IAASEnvironment, ip, password, ca string) error {
	var (
		stemcell string
		err      error
	)
	iaas := config.IAASCheck()
	switch iaas {
	case awsConst:
		stemcell, err = config.ConfigureConcourseStemcell(resource.AWSReleaseVersions)

	case gcpConst:
		stemcell, err = config.ConfigureConcourseStemcell(resource.GCPReleaseVersions)
	}
	if err != nil {
		return err
	}

	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)
	cmd := c.execCmd(c.boshPath, "--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "upload-stemcell", stemcell)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// Recreate runs BOSH recreate
func (c *BOSHCLI) Recreate(config IAASEnvironment, ip, password, ca string) error {
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)
	cmd := c.execCmd(c.boshPath, "--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "--deployment", "concourse", "recreate")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (c *BOSHCLI) DeleteEnv(store Store, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) error {
	return c.xEnv("delete-env", store, config, password, cert, key, ca, tags)
}

func (c *BOSHCLI) CreateEnv(store Store, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) error {

	return c.xEnv("create-env", store, config, password, cert, key, ca, tags)
}

func (c *BOSHCLI) RunAuthenticatedCommand(action, ip, password, ca string, detach bool, stdout io.Writer, flags ...string) error {
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)

	authFlags := []string{"--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "--deployment", "concourse", action}
	flags = append(authFlags, flags...)
	if detach {
		return c.detachedBoshCommand(stdout, flags...)
	}
	return c.boshCommand(stdout, flags...)
}

func (c *BOSHCLI) boshCommand(stdout io.Writer, flags ...string) error {
	cmd := c.execCmd(c.boshPath, flags...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = stdout
	return cmd.Run()
}

func (c *BOSHCLI) detachedBoshCommand(stdout io.Writer, flags ...string) error {
	cmd := c.execCmd(c.boshPath, flags...)
	cmd.Stderr = os.Stderr

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
		if _, err := stdout.Write([]byte(fmt.Sprintf("%s\n", text))); err != nil {
			return err
		}
		if strings.Contains(text, "Preparing deployment") {
			stdout.Write([]byte("Task started, detaching output\n"))
			return nil
		}
	}

	return fmt.Errorf("Didn't detect successful task start in BOSH comand: bosh-cli %s", strings.Join(flags, " "))
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
