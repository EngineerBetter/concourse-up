package bincache_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/EngineerBetter/concourse-up/util/bincache"
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

// check that download handles zip files
func TestDownloadZip(t *testing.T) {
	file, _ := ioutil.TempFile("", "*fAKETerraformBinary.sh")
	file.WriteString("#!/bin/bash\necho HELLO")
	file.Close()
	defer os.Remove(file.Name())

	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)
	data, _ := ioutil.ReadFile(file.Name())
	f, _ := writer.Create(file.Name())
	f.Write([]byte(data))

	writer.Close()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", file.Name()))
		w.Write(buf.Bytes())
	}))

	path, err := bincache.Download(s.URL)
	require.NoError(t, err)
	defer os.Remove(path)
	out, err := exec.Command(path).Output()
	require.NoError(t, err)
	require.Equal(t, "HELLO\n", string(out))
	s.Close()
}
