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

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/util"
)

// ConcourseUpVersion is a compile-time variable set with -ldflags
var ConcourseUpVersion = "COMPILE_TIME_VARIABLE_fly_concourse_up_version"

// IClient represents an interface for a client
type IClient interface {
	CanConnect() (bool, error)
	SetDefaultPipeline(deployArgs *config.DeployArgs, config *config.Config, allowFlyVersionDiscrepancy bool) error
	Cleanup() error
}

// Client represents a low-level wrapper for fly
type Client struct {
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
func New(creds Credentials, stdout, stderr io.Writer, versionFile []byte) (IClient, error) {
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

	return &Client{
		tempDir,
		creds,
		stdout,
		stderr,
		versionFile,
	}, nil
}

// CanConnect returns true if it can connect to the concourse
func (client *Client) CanConnect() (bool, error) {
	cmd := exec.Command(
		client.tempDir.Path("fly"),
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
func (client *Client) SetDefaultPipeline(deployArgs *config.DeployArgs, config *config.Config, allowFlyVersionDiscrepancy bool) error {
	if err := client.login(); err != nil {
		return err
	}

	if allowFlyVersionDiscrepancy {
		if err := client.sync(); err != nil {
			return err
		}
	}

	pipelinePath := client.tempDir.Path("default-pipeline.yml")
	pipelineName := "concourse-up-self-update"

	if err := client.writePipelineConfig(pipelinePath, deployArgs, config); err != nil {
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

func (client *Client) writePipelineConfig(pipelinePath string, deployArgs *config.DeployArgs, config *config.Config) error {
	fileHandler, err := os.Create(pipelinePath)
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	params, err := client.buildDefaultPipelineParams(deployArgs, config)
	if err != nil {
		return err
	}

	pipelineConfig, err := util.RenderTemplate(defaultPipelineTemplate, params)
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
	cmd := exec.Command(client.tempDir.Path("fly"), args...)
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
	case "windows":
		return x["fly"]["windows"], nil
	default:
		return "", fmt.Errorf("unknown os: `%s`", runtime.GOOS)
	}
}

func (client *Client) buildDefaultPipelineParams(deployArgs *config.DeployArgs, config *config.Config) (*defaultPipelineParams, error) {
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if awsAccessKeyID == "" {
		return nil, errors.New("env var AWS_ACCESS_KEY_ID not found")
	}

	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretAccessKey == "" {
		return nil, errors.New("env var AWS_SECRET_ACCESS_KEY not found")
	}

	return &defaultPipelineParams{
		AWSAccessKeyID:     awsAccessKeyID,
		AWSSecretAccessKey: awsSecretAccessKey,
		Deployment:         strings.TrimPrefix(config.Deployment, "concourse-up-"),
		FlagAWSRegion:      deployArgs.AWSRegion,
		FlagDomain:         deployArgs.Domain,
		FlagTLSCert:        deployArgs.TLSCert,
		FlagTLSKey:         deployArgs.TLSKey,
		FlagWebSize:        deployArgs.WebSize,
		FlagWorkerSize:     deployArgs.WorkerSize,
		FlagWorkers:        deployArgs.WorkerCount,
		ConcourseUpVersion: ConcourseUpVersion,
	}, nil
}

type defaultPipelineParams struct {
	AWSAccessKeyID     string
	AWSDefaultRegion   string
	AWSSecretAccessKey string
	Deployment         string
	FlagAWSRegion      string
	FlagDomain         string
	FlagTLSCert        string
	FlagTLSKey         string
	FlagWebSize        string
	FlagWorkerSize     string
	FlagWorkers        int
	ConcourseUpVersion string
}

// Indent is a helper function to indent the field a given number of spaces
func (params defaultPipelineParams) Indent(countStr, field string) string {
	return util.Indent(countStr, field)
}

const defaultPipelineTemplate = `
---
resources:
- name: concourse-up-release
  type: github-release
  source:
    user: engineerbetter
    repository: concourse-up
    pre_release: true
- name: every-month
  type: time
  source: {interval: 730h}

jobs:
- name: self-update
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    trigger: true
  - task: update
    params:
      AWS_REGION: "<% .FlagAWSRegion %>"
      DOMAIN: "<% .FlagDomain %>"
      TLS_CERT: |-
        <% .Indent "8" .FlagTLSCert %>
      TLS_KEY: |-
        <% .Indent "8" .FlagTLSKey %>
      WORKERS: "<% .FlagWorkers %>"
      WORKER_SIZE: "<% .FlagWorkerSize %>"
      WEB_SIZE: "<% .FlagWebSize %>"
      DEPLOYMENT: "<% .Deployment %>"
      AWS_ACCESS_KEY_ID: "<% .AWSAccessKeyID %>"
      AWS_SECRET_ACCESS_KEY: "<% .AWSSecretAccessKey %>"
      SELF_UPDATE: true
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/cup-image
      inputs:
      - name: concourse-up-release
      run:
        path: bash
        args:
        - -c
        - |
          set -eux

          cd concourse-up-release
          chmod +x concourse-up-linux-amd64
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
- name: renew-cert
  serial_groups: [cup]
  serial: true
  plan:
  - get: concourse-up-release
    version: {tag: <% .ConcourseUpVersion %> }
  - get: every-month
    trigger: true
  - task: update
    params:
      AWS_REGION: "<% .FlagAWSRegion %>"
      DOMAIN: "<% .FlagDomain %>"
      TLS_CERT: |-
        <% .Indent "8" .FlagTLSCert %>
      TLS_KEY: |-
        <% .Indent "8" .FlagTLSKey %>
      WORKERS: "<% .FlagWorkers %>"
      WORKER_SIZE: "<% .FlagWorkerSize %>"
      WEB_SIZE: "<% .FlagWebSize %>"
      DEPLOYMENT: "<% .Deployment %>"
      AWS_ACCESS_KEY_ID: "<% .AWSAccessKeyID %>"
      AWS_SECRET_ACCESS_KEY: "<% .AWSSecretAccessKey %>"
      SELF_UPDATE: true
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/cup-image
      inputs:
      - name: concourse-up-release
      run:
        path: bash
        args:
        - -c
        - |
          set -eux

          cd concourse-up-release
          chmod +x concourse-up-linux-amd64
          ./concourse-up-linux-amd64 deploy $DEPLOYMENT
`
