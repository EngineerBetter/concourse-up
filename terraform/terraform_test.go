package terraform_test

import (
	"bytes"
	"github.com/EngineerBetter/concourse-up/iaas"
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
	mockCLIent, err := terraform.New(iaas.AWS, terraform.FakeExec(e.Cmd()))
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
	err = mockCLIent.Apply(config)
	require.NoError(t, err)
}

func TestCLI_ApplyPlan(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockCLIent, err := terraform.New(iaas.AWS, terraform.FakeExec(e.Cmd()))
	require.NoError(t, err)

	config := &mockTerraformInputVars{}

	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "init")

	})
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "terraform", command)
		require.Equal(t, args[0], "apply")
	})
	err = mockCLIent.Apply(config)
	require.NoError(t, err)
}

func TestCLI_Destroy(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	mockCLIent, err := terraform.New(iaas.AWS, terraform.FakeExec(e.Cmd()))
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
