package boshenv_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/EngineerBetter/concourse-up/bosh/internal/boshenv"
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

type mockStore map[string][]byte

func (s mockStore) Set(key string, value []byte) error {
	s[key] = value
	return nil
}

func (s mockStore) Get(key string) ([]byte, error) {
	return s[key], nil
}

type mockIAASConfig struct {
	f func(string) string
}

func (c mockIAASConfig) IAASCheck() string {
	return "AWS"
}
func (c mockIAASConfig) ConfigureDirectorManifestCPI(m string) (string, error) {
	return c.f(m), nil
}

func (c mockIAASConfig) ConfigureDirectorCloudConfig(m string) (string, error) {
	return c.f(m), nil
}

func (c mockIAASConfig) ConfigureConcourseStemcell(v string) (string, error) {
	return c.f(v), nil
}

func TestBOSHCLI_CreateEnv(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c, err := boshenv.New(boshenv.FakeExec(e.Cmd()))
	require.NoError(t, err)
	store := make(mockStore)
	config := mockIAASConfig{
		f: func(s string) string {
			return s
		},
	}
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)
		require.Equal(t, "create-env", args[0])
		s := strings.TrimPrefix(args[1], "--state=")
		expectPathNotToExistButBeWriteable(t, s)
		v := strings.TrimPrefix(args[2], "--vars-store=")
		expectPathNotToExistButBeWriteable(t, v)
	})
	err = c.CreateEnv(store, config, "password", "cert", "key", "ca", map[string]string{})
	require.NoError(t, err)
}

func expectPathNotToExistButBeWriteable(t testing.TB, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected path %q not to exist", path)
	}
	f, err := os.Create(path)
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()
}

func TestBOSHCLI_UpdateCloudConfig(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c, err := boshenv.New(boshenv.FakeExec(e.Cmd()))
	require.NoError(t, err)
	config := mockIAASConfig{
		f: func(s string) string {
			return s
		},
	}
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)

		require.Equal(t, "--non-interactive", args[0])
		require.Equal(t, "--environment", args[1])
		require.Equal(t, "https://ip", args[2])
		require.Equal(t, "--client-secret", args[7])
		require.Equal(t, "password", args[8])
		require.Equal(t, "update-cloud-config", args[9])
	})
	err = c.UpdateCloudConfig(config, "ip", "password", "ca")
	require.NoError(t, err)
}

func TestBOSHCLI_UploadConcourseStemcell(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c, err := boshenv.New(boshenv.FakeExec(e.Cmd()))
	require.NoError(t, err)
	config := mockIAASConfig{
		f: func(s string) string {
			return s
		},
	}
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)

		require.Equal(t, "--non-interactive", args[0])
		require.Equal(t, "--environment", args[1])
		require.Equal(t, "https://ip", args[2])
		require.Equal(t, "--client-secret", args[7])
		require.Equal(t, "password", args[8])
		require.Equal(t, "upload-stemcell", args[9])
	})
	err = c.UploadConcourseStemcell(config, "ip", "password", "ca")
	require.NoError(t, err)

}
