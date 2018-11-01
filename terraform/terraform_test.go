package terraform_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/EngineerBetter/concourse-up/terraform/internal/aws"
	"github.com/EngineerBetter/concourse-up/terraform/internal/gcp"

	"github.com/EngineerBetter/concourse-up/internal/fakeexec"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/stretchr/testify/require"
)

type mockTerraformInputVars struct{}
type mockIAASMetadata struct{}

func (mockIAASMD *mockIAASMetadata) AssertValid() error {
	return nil
}
func (mockIAASMD *mockIAASMetadata) Init(buffer *bytes.Buffer) error {
	return nil
}
func (mockIAASMD *mockIAASMetadata) Get(key string) (string, error) {
	return "", nil
}

func (mockInputVars *mockTerraformInputVars) ConfigureTerraform(terraformContents string) (string, error) {
	return "", nil
}

func (mockInputVars *mockTerraformInputVars) Build(data map[string]interface{}) error {
	return nil
}
func TestCLI_Apply(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockCLIent, err := terraform.New(terraform.FakeExec(e.Cmd()))
	require.NoError(t, err)

	config := &mockTerraformInputVars{}

	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "init")

	})
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "apply")
		require.Equal(t, args[1], "-input=false")
		require.Equal(t, args[2], "-auto-approve")

	})
	err = mockCLIent.Apply(config, false)
	require.NoError(t, err)
}

func TestCLI_ApplyPlan(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockCLIent, err := terraform.New(terraform.FakeExec(e.Cmd()))
	require.NoError(t, err)

	config := &mockTerraformInputVars{}

	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "init")

	})
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "plan")
	})
	err = mockCLIent.Apply(config, true)
	require.NoError(t, err)
}

func TestCLI_Destroy(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockCLIent, err := terraform.New(terraform.FakeExec(e.Cmd()))
	require.NoError(t, err)

	config := &mockTerraformInputVars{}

	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "init")

	})
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "destroy")
		require.Equal(t, args[1], "-auto-approve")

	})
	err = mockCLIent.Destroy(config)
	require.NoError(t, err)
}

func TestCLI_BuildOutput(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockCLIent, err := terraform.New(terraform.FakeExec(e.Cmd()))
	require.NoError(t, err)

	config := &mockTerraformInputVars{}
	metadata := &mockIAASMetadata{}

	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "init")

	})
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "output")
		require.Equal(t, args[1], "-json")

	})
	err = mockCLIent.BuildOutput(config, metadata)
	require.NoError(t, err)
}

func TestCLI_IAAS(t *testing.T) {
	tests := []struct {
		name             string
		args             string
		wantInputVars    terraform.InputVars
		wantIAASMetadata terraform.IAASMetadata
		wantErr          bool
	}{
		{
			name:             "return GCP provider hooks for GCP",
			args:             "GCP",
			wantInputVars:    &gcp.InputVars{},
			wantIAASMetadata: &gcp.Metadata{},
			wantErr:          false,
		},
		{
			name:             "return AWS provider hooks for AWS",
			args:             "AWS",
			wantInputVars:    &aws.InputVars{},
			wantIAASMetadata: &aws.Metadata{},
			wantErr:          false,
		},
		{
			name:             "return null provider hooks for unknown provider",
			args:             "aProvider",
			wantInputVars:    &terraform.NullInputVars{},
			wantIAASMetadata: &terraform.NullMetadata{},
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := fakeexec.New(t)
			defer e.Finish()
			c, err := terraform.New(terraform.FakeExec(e.Cmd()))
			require.NoError(t, err)

			gotInputVars, gotIAASMetadata, err := c.IAAS(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("CLI.IAAS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInputVars, tt.wantInputVars) {
				t.Errorf("CLI.IAAS() gotInputVars = %v, wantInputVars %v", gotInputVars, tt.wantInputVars)
			}
			if !reflect.DeepEqual(gotIAASMetadata, tt.wantIAASMetadata) {
				t.Errorf("CLI.IAAS() gotIAASMetadata = %v, wantIAASMetadata %v", gotIAASMetadata, tt.wantIAASMetadata)
			}
		})
	}
}
