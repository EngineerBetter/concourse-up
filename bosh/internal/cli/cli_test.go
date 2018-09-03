package cli_test

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/EngineerBetter/concourse-up/bosh/internal/cli"
	"github.com/EngineerBetter/concourse-up/internal/fakeexec"
	"github.com/stretchr/testify/require"
)

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Print(os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}
func TestBOSHCLI_Interpolate(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c, err := cli.New(cli.FakeExec(e.Cmd()))
	require.NoError(t, err)
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)
		require.Equal(t, "interpolate", args[0])
		sort.Strings(args[1:2])
		require.Equal(t, `-v=foo="bar"`, args[1])
		require.Equal(t, `-v=baz=["alpha","beta"]`, args[2])
		f := strings.TrimPrefix(args[3], "-o=")
		require.FileExists(t, f)
		require.Equal(t, "-", args[4])
	}).Outputs("my manifest")
	result, err := cli.Interpolate("", []string{
		"my ops file",
	}, map[string]interface{}{
		"foo": "bar",
		"baz": []string{"alpha", "beta"},
	})
	require.NoError(t, err)
	require.Equal(t, "my manifest", result)
}

type mockStore map[string][]byte

func (s mockStore) Set(key string, value []byte) error {
	s[key] = value
	return nil
}

func (s mockStore) Get(key string) ([]byte, error) {
	return s[key], nil
}

type mockIAASConfig struct {
	operations []string
	vars       map[string]interface{}
}

func (c mockIAASConfig) Operations() []string {
	return c.operations
}

func (c mockIAASConfig) Vars() map[string]interface{} {
	return c.vars
}

func TestBOSHCLI_CreateEnv(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c, err := cli.New(cli.FakeExec(e.Cmd()))
	require.NoError(t, err)
	store := make(mockStore)
	config := mockIAASConfig{
		operations: []string{
			"my operation",
		},
		vars: map[string]interface{}{
			"foo": "bar",
		},
	}
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)
		require.Equal(t, "interpolate", args[0])
		require.Equal(t, `-v=foo="bar"`, args[1])
		f := strings.TrimPrefix(args[2], "-o=")
		require.FileExists(t, f)
		require.Equal(t, "-", args[3])
	}).Outputs("my manifest")
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)
		require.Equal(t, "create-env", args[0])
		s := strings.TrimPrefix(args[1], "--state=")
		require.FileExists(t, s)
		v := strings.TrimPrefix(args[2], "--vars-store=")
		require.FileExists(t, v)
		require.Equal(t, "-", args[3])
	})
	err = c.CreateEnv("my manifest", store, config)
	require.NoError(t, err)
}
