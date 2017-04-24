package bosh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

func (client *Client) directorSSH() (*ssh.Client, error) {
	key, err := ssh.ParsePrivateKey([]byte(client.config.PrivateKey))
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: "vcap",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	sshClient, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:22", client.metadata.DirectorPublicIP.Value),
		sshConfig,
	)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}
