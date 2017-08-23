package fly

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/EngineerBetter/concourse-up/util"
)

// DarwinBinaryURL is a compile-time variable set with -ldflags
var DarwinBinaryURL = "COMPILE_TIME_VARIABLE_fly_darwin_binary_url"

// LinuxBinaryURL is a compile-time variable set with -ldflags
var LinuxBinaryURL = "COMPILE_TIME_VARIABLE_fly_linux_binary_url"

// WindowsBinaryURL is a compile-time variable set with -ldflags
var WindowsBinaryURL = "COMPILE_TIME_VARIABLE_fly_windows_binary_url"

// IClient represents an interface for a client
type IClient interface {
	SetDefaultPipeline() error
	Cleanup() error
}

// Client represents a low-level wrapper for fly
type Client struct {
	tempDir *util.TempDir
	creds   Credentials
	stdout  io.Writer
	stderr  io.Writer
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
func New(creds Credentials, stdout, stderr io.Writer) (IClient, error) {
	tempDir, err := util.NewTempDir()
	if err != nil {
		return nil, err
	}

	fileHandler, err := os.Create(tempDir.Path("fly"))
	if err != nil {
		return nil, err
	}
	defer fileHandler.Close()

	url, err := getFlyURL()
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
	}, nil
}

// SetDefaultPipeline sets the default pipeline against a given concourse
func (client *Client) SetDefaultPipeline() error {
	if err := client.login(); err != nil {
		return err
	}

	pipelinePath := client.tempDir.Path("default-pipeline.yml")
	pipelineName := "hello-world"

	fileHandler, err := os.Create(pipelinePath)
	if err != nil {
		return err
	}
	defer fileHandler.Close()

	if _, err := fileHandler.WriteString(defaultPipeline); err != nil {
		return err
	}

	if err := fileHandler.Sync(); err != nil {
		return err
	}

	if err := client.run("set-pipeline", "--pipeline", pipelineName, "--config", pipelinePath, "--non-interactive"); err != nil {
		return err
	}

	if err := os.Remove(pipelinePath); err != nil {
		return err
	}

	if err := client.run("unpause-pipeline", "--pipeline", pipelineName); err != nil {
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

	client.stdout.Write([]byte("Waiting for Concourse ATC to start... \n"))

	for i := 0; i < attempts; i++ {
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
			return nil
		}

		stderrBytes, err := ioutil.ReadAll(stderr)
		if err != nil {
			return err
		}

		if strings.Contains(string(stderrBytes), "could not reach the Concourse server") {
			time.Sleep(time.Second)
			continue
		}

		// if there is a legitimate error, copy it to stderr for debugging
		if _, err := client.stderr.Write(stderrBytes); err != nil {
			return err
		}

		return runErr
	}

	return fmt.Errorf("failed to log in to %s after %d attempts", client.creds.API, attempts)
}

func (client *Client) run(args ...string) error {
	args = append([]string{"--target", client.creds.Target}, args...)
	cmd := exec.Command(client.tempDir.Path("fly"), args...)
	cmd.Stdout = client.stdout
	cmd.Stderr = client.stderr
	return cmd.Run()
}

func getFlyURL() (string, error) {
	os := runtime.GOOS
	if os == "darwin" {
		return DarwinBinaryURL, nil
	} else if os == "linux" {
		return LinuxBinaryURL, nil
	} else if os == "windows" {
		return WindowsBinaryURL, nil
	}
	return "", fmt.Errorf("unknown os: `%s`", os)
}

const defaultPipeline = `
jobs:
- name: hello-world
  plan:
  - task: say-hello
    config:
      platform: linux
      image_resource:
        type: docker-image
        source: {repository: ubuntu}
      run:
        path: echo
        args: ["Hello, world!"]
`
