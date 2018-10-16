package bincache_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/EngineerBetter/concourse-up/bosh/internal/bincache"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "#!/bin/bash\necho hi")
	}))
	defer s.Close()
	path, err := bincache.Download(s.URL)
	require.NoError(t, err)
	defer os.Remove(path)
	out, err := exec.Command(path).Output()
	require.NoError(t, err)
	require.Equal(t, "hi\n", string(out))
	// check download does not happen if file already exists
	s.Close()
	path1, err := bincache.Download(s.URL)
	require.NoError(t, err)
	require.Equal(t, path, path1)
}
