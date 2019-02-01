package fly

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/EngineerBetter/concourse-up/iaas"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/util"
)

// ConcourseUpVersion is a compile-time variable set with -ldflags
var ConcourseUpVersion = "COMPILE_TIME_VARIABLE_fly_concourse_up_version"

// IClient represents an interface for a client
type IClient interface {
	CanConnect() (bool, error)
	SetDefaultPipeline(config config.Config, allowFlyVersionDiscrepancy bool) error
	Cleanup() error
}

// Client represents a low-level wrapper for fly
type Client struct {
	pipeline    Pipeline
	tempDir     *util.TempDir
	creds       Credentials
	stdout      io.Writer
	stderr      io.Writer
	versionFile []byte
}

// Credentials represents credentials needed to connect to concourse
type Credentials struct {
	Target   string
	API      string
	Username string
	Password string
	CACert   string
}

// New returns a new fly client
func New(provider iaas.Provider, creds Credentials, stdout, stderr io.Writer, versionFile []byte) (IClient, error) {
	tempDir, err := util.NewTempDir()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.Create(tempDir.Path("fly"))
	if err != nil {
		return nil, err
	}
	defer fileHandler.Close()

	url, err := getFlyURL(versionFile)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(fileHandler, resp.Body); err != nil {
		return nil, err
	}

	if err := fileHandler.Sync(); err != nil {
		return nil, err
	}

	if err := os.Chmod(fileHandler.Name(), 0700); err != nil {
		return nil, err
	}

	var pipeline Pipeline

	switch provider.IAAS() {
	case "AWS":
		pipeline = NewAWSPipeline()
	case "GCP":
		GCPCreds, err := provider.Attr("credentials_path")
		if err != nil {
			return nil, err
		}
		pipeline, err = NewGCPPipeline(GCPCreds)
		if err != nil {
			return nil, errors.New("fly.go: failed to read credentials file")
		}
	default:
		return nil, errors.New("fly.go: IAAS not recognised")

	}
	return &Client{
		pipeline,
		tempDir,
		creds,
		stdout,
		stderr,
		versionFile,
	}, nil
}

var (
	execCommand = exec.Command
)

func (client *Client) runFly(args ...string) *exec.Cmd {
	return execCommand(client.tempDir.Path("fly"), args...)
}

// CanConnect returns true if it can connect to the concourse
func (client *Client) CanConnect() (bool, error) {
	cmd := client.runFly(
		"--target",
		client.creds.Target,
		"login",
		"--insecure",
		"--concourse-url",
		client.creds.API,
		"--username",
		client.creds.Username,
		"--password",
		client.creds.Password,
	)

	stderr := bytes.NewBuffer(nil)
	cmd.Stdout = client.stdout
	cmd.Stderr = stderr

	runErr := cmd.Run()
	if runErr == nil {
		return true, nil
	}

	stderrBytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(stderrBytes), "could not reach the Concourse server") {
		return false, nil
	}

	// if there is a legitimate error, copy it to stderr for debugging
	if _, err := client.stderr.Write(stderrBytes); err != nil {
		return false, err
	}

	return false, runErr
}

// SetDefaultPipeline sets the default pipeline against a given concourse
func (client *Client) SetDefaultPipeline(config config.Config, allowFlyVersionDiscrepancy bool) error {
	if err := client.login(); err != nil {
		return err
	}

	if allowFlyVersionDiscrepancy {
		if err := client.sync(); err != nil {
			return err
		}
		if err := client.login(); err != nil {
			return err
		}
	}

	pipelinePath := client.tempDir.Path("default-pipeline.yml")
	pipelineName := "concourse-up-self-update"

	if err := client.writePipelineConfig(pipelinePath, config); err != nil {
		return err
	}

	if err := client.run("set-pipeline", "--pipeline", pipelineName, "--config", pipelinePath, "--non-interactive"); err != nil {
		return err
	}

	if err := os.Remove(pipelinePath); err != nil {
		return err
	}

	if err := client.run("pause-job", "--job", pipelineName+"/self-update"); err != nil {
		return err
	}

	return client.run("unpause-pipeline", "--pipeline", pipelineName)
}

func (client *Client) writePipelineConfig(pipelinePath string, config config.Config) error {
	fileHandler, err := os.Create(pipelinePath)
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	params, err := client.pipeline.BuildPipelineParams(config.Deployment, config.Namespace, config.Region, config.Domain)
	if err != nil {
		return err
	}
	pipelineTemplate := client.pipeline.GetConfigTemplate()
	pipelineConfig, err := util.RenderTemplate("self-update pipeline", pipelineTemplate, params)
	if err != nil {
		return err
	}

	if _, err := fileHandler.Write(pipelineConfig); err != nil {
		return err
	}

	if err := fileHandler.Sync(); err != nil {
		return err
	}

	return nil
}

// Cleanup removes tempfiles
func (client *Client) Cleanup() error {
	return client.tempDir.Cleanup()
}

func (client *Client) login() error {
	attempts := 50
	secondsBetweenAttempts := 4

	if _, err := client.stdout.Write([]byte("Waiting for Concourse ATC to start... \n")); err != nil {
		return err
	}

	for i := 0; i < attempts; i++ {
		canConnect, err := client.CanConnect()
		if err != nil {
			return err
		}
		if canConnect {
			return nil
		}

		time.Sleep(time.Second * time.Duration(secondsBetweenAttempts))
	}

	return fmt.Errorf("failed to log in to %s after %d seconds", client.creds.API, attempts*secondsBetweenAttempts)
}

func (client *Client) sync() error {
	return client.run("sync")
}

func (client *Client) run(args ...string) error {
	args = append([]string{"--target", client.creds.Target}, args...)
	cmd := client.runFly(args...)
	cmd.Stdout = client.stdout
	cmd.Stderr = client.stderr
	return cmd.Run()
}

func getFlyURL(versionFile []byte) (string, error) {
	var x map[string]map[string]string
	err := json.Unmarshal(versionFile, &x)
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return x["fly"]["mac"], nil
	case "linux":
		return x["fly"]["linux"], nil
	default:
		return "", fmt.Errorf("unknown os: `%s`", runtime.GOOS)
	}
}
