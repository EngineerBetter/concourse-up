package terraform_test

import (
	"bytes"
	"github.com/EngineerBetter/concourse-up/iaas"
	"reflect"
	"testing"

	"github.com/EngineerBetter/concourse-up/internal/fakeexec"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/stretchr/testify/require"
)

type mockTerraformInputVars struct{}
type mockOutputs struct{}

func (mockIAASMD *mockOutputs) AssertValid() error {
	return nil
}
func (mockIAASMD *mockOutputs) Init(buffer *bytes.Buffer) error {
	return nil
}
func (mockIAASMD *mockOutputs) Get(key string) (string, error) {
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
	mockCLIent.IAAS(iaas.AWS)

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
	mockCLIent.IAAS(iaas.AWS)

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
	mockCLIent.IAAS(iaas.AWS)

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
	mockCLIent.IAAS(iaas.AWS)

	config := &mockTerraformInputVars{}
	outputs := &mockOutputs{}

	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "init")

	})
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "output")
		require.Equal(t, args[1], "-json")

	})
	err = mockCLIent.BuildOutput(config, outputs)
	require.NoError(t, err)
}

func TestCLI_IAAS(t *testing.T) {
	tests := []struct {
		name             string
		args             iaas.Name
		wantInputVars    terraform.InputVars
		wantIAASMetadata terraform.Outputs
		wantErr          bool
	}{
		{
			name:             "return GCP provider hooks for GCP",
			args:             iaas.GCP,
			wantIAASMetadata: &terraform.GCPOutputs{},
			wantErr:          false,
		},
		{
			name:             "return AWS provider hooks for AWS",
			args:             iaas.AWS,
			wantIAASMetadata: &terraform.AWSOutputs{},
			wantErr:          false,
		},
		{
			name:             "return null provider hooks for unknown provider",
			args:             iaas.Unknown,
			wantIAASMetadata: &terraform.NullOutputs{},
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := fakeexec.New(t)
			defer e.Finish()
			c, err := terraform.New(terraform.FakeExec(e.Cmd()))
			require.NoError(t, err)

			gotIAASMetadata, err := c.IAAS(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("CLI.IAAS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIAASMetadata, tt.wantIAASMetadata) {
				t.Errorf("CLI.IAAS() gotIAASMetadata = %v, wantIAASMetadata %v", gotIAASMetadata, tt.wantIAASMetadata)
			}
		})
	}
}
