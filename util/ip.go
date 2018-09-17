package util

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// FindUserIP gets the user's public IP by querying whatismyip.akamai.com
func FindUserIP() (string, error) {
	const retries = 10
	for i := 0; i < retries; i++ {
		resp, err := http.Get("http://whatismyip.akamai.com")
		if err != nil {
			return "", err
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		ip := strings.TrimSpace(string(bytes))
		if ip != "" {
			return ip, nil
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return "", errors.New("timed out getting user IP")
}
