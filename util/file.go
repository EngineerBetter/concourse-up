package util

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Path expands ~ to home directory in file path
func Path(path string) (string, error) {
	var err error
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path[:2] == "~/" {
		path = filepath.Join(dir, path[2:])
	}
	if strings.Contains(path, "~") {
		err = errors.New("Invalid Path")
	}
	return path, err
}

// AssertDirExists checks if the dir `path` exists and creates it if it doesn't
func AssertDirExists(path string) error {
	var err error
	fmt.Printf("Asserting `%s` is present\n", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Creating `%s`\n", path)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
	return err
}

// AssertFileExists checks if file `path` exists and creates it if it doesn't
func AssertFileExists(path string) error {
	var err error
	fmt.Printf("Asserting `%s` is present\n", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Creating `%s`\n", path)
		configFile, err := os.Create(path)
		if err != nil {
			fmt.Println("Error:", err)
		}
		configFile.Close()
	}
	return err
}
