package terraform_test

import (
	"bytes"
	"testing"

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
func TestTerraformCLI_Apply(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockTerraformClient, err := terraform.New(terraform.FakeExec(e.Cmd()))
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
	err = mockTerraformClient.Apply(config, false)
	require.NoError(t, err)
}

func TestTerraformCLI_ApplyPlan(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockTerraformClient, err := terraform.New(terraform.FakeExec(e.Cmd()))
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
	err = mockTerraformClient.Apply(config, true)
	require.NoError(t, err)
}

func TestTerraformCLI_Destroy(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockTerraformClient, err := terraform.New(terraform.FakeExec(e.Cmd()))
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
	err = mockTerraformClient.Destroy(config)
	require.NoError(t, err)
}

func TestTerraformCLI_BuildOutput(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockTerraformClient, err := terraform.New(terraform.FakeExec(e.Cmd()))
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
	err = mockTerraformClient.BuildOutput(config, metadata)
	require.NoError(t, err)
}
