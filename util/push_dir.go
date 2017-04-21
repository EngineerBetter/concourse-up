package util

import (
	"os"
	"path/filepath"
)

// PushDir runs callback in given directory. Propagates any errors returned by callback
func PushDir(dir string, callback func() error) error {
	initialDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	if err = os.Chdir(dir); err != nil {
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
