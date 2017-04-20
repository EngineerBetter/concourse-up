package util

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// FindUserIP gets the user's public IP by querying whatismyip.akamai.com
func FindUserIP() (string, error) {
	resp, err := http.Get("http://whatismyip.akamai.com")
	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}
