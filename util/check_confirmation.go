package util

import (
	"fmt"
	"io"
	"strings"
)

// CheckConfirmation prompts the user for confirmation and returns true IFF the user responds with 'yes'
func CheckConfirmation(stdin io.Reader, stdout io.Writer, name string) (bool, error) {
	var response string
	fmt.Fprintf(stdout, "Are you sure you want to destroy %s?\nThis cannot be undone. [yes/no]: ", name)

	_, err := fmt.Fscan(stdin, &response)
	if err != nil {
		return false, err
	}
	response = strings.TrimSpace(response)
	response = strings.ToLower(response)
	if response == "yes" {
		return true, nil
	} else if response == "no" {
		return false, nil
	}

	return false, fmt.Errorf("Input not recognized: `%s`", response)
}
