package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// TempDir represents a temporary directory to store files
type TempDir struct {
	path string
}

// NewTempDir returns a new temporary directory
func NewTempDir() (*TempDir, error) {
	path, err := ioutil.TempDir("", "concourse-up")
	if err != nil {
		return nil, err
	}

	return &TempDir{
		path: path,
	}, nil
}

// Save writes the given file into the tempDir
func (tempDir *TempDir) Save(filename string, contents []byte) (string, error) {
	path := filepath.Join(tempDir.path, filename)
	if err := ioutil.WriteFile(path, contents, 0700); err != nil {
		return "", err
	}

	return path, nil
}

// Path returns a path for the given file into the tempDir
func (tempDir *TempDir) Path(filename string) string {
	return filepath.Join(tempDir.path, filename)
}

// Cleanup deletes the tempDir
func (tempDir *TempDir) Cleanup() error {
	return os.RemoveAll(tempDir.path)
}

// PushDir runs the function in the tempDir
func (tempDir *TempDir) PushDir(callback func() error) error {
	initialDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	if err = os.Chdir(tempDir.path); err != nil {
		return err
	}

	if err = callback(); err != nil {
		return err
	}

	if err = os.Chdir(initialDir); err != nil {
		return err
	}

	return nil
}
