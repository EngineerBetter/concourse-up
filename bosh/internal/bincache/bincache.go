package bincache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	_, err = io.Copy(f, resp.Body)
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
