package fly

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/EngineerBetter/concourse-up/iaas"

	"github.com/EngineerBetter/concourse-up/util"
)

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// some code here to check arguments perhaps?
	fmt.Fprintf(os.Stdout, os.Getenv("TEST_HELPER_OUTPUT"))
	os.Exit(0)
}

func TestClient_runFly(t *testing.T) {
	type fields struct {
		pipeline    Pipeline
		provider    iaas.Provider
		tempDir     *util.TempDir
		creds       Credentials
		stdout      io.Writer
		stderr      io.Writer
		versionFile []byte
	}
	type args struct {
		args []string
	}
	tmpDir, _ := util.NewTempDir()
	tests := []struct {
		name      string
		fields    fields
		args      args
		cmdOutput string
		want      string
	}{{
		name: "sync",
		fields: fields{
			pipeline:    nil,
			provider:    nil,
			tempDir:     tmpDir,
			creds:       Credentials{},
			stdout:      nil,
			stderr:      nil,
			versionFile: nil,
		},
		args:      args{[]string{"sync"}},
		cmdOutput: "syncing",
		want:      "syncing",
	}}

	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				provider:    tt.fields.provider,
				tempDir:     tt.fields.tempDir,
				creds:       tt.fields.creds,
				stdout:      tt.fields.stdout,
				stderr:      tt.fields.stderr,
				versionFile: tt.fields.versionFile,
			}
			os.Setenv("TEST_HELPER_OUTPUT", tt.cmdOutput)

			if got, _ := client.runFly(tt.args.args...).CombinedOutput(); !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("Client.runFly() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
