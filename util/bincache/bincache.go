package bincache

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Download a file from url and check with a sha1
func Download(url string) (string, error) {
	path, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	path = filepath.Join(path, "concourse-up", "bin")
	err = os.MkdirAll(path, 0700)
	if err != nil {
		return "", err
	}
	path = filepath.Join(path, hash(url))
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		return path, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return "", err
	}
	defer f.Close()
	resp, err := http.Get(url)
	if err != nil {
		os.Remove(path)
		return "", err
	}
	defer resp.Body.Close()

	var closer io.ReadCloser

	if strings.HasSuffix(url, ".zip") {
		body, errz := ioutil.ReadAll(resp.Body)
		if errz != nil {
			return "", errz
		}

		r, errz := zip.NewReader(bytes.NewReader(body), resp.ContentLength)
		if errz != nil {
			return "", errz
		}
		firstFile, errz := r.File[0].Open()
		if errz != nil {
			return "", errz
		}
		closer = firstFile
	} else {
		closer = resp.Body
	}
	_, err = io.Copy(f, closer)
	if err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
