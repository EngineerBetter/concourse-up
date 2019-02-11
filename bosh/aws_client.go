package bosh

import (
	"fmt"
	"io"
	"net"

	"github.com/EngineerBetter/concourse-up/config"
	"github.com/EngineerBetter/concourse-up/director"
	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/terraform"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

//AWSClient is an AWS specific implementation of IClient
type AWSClient struct {
	config   config.Config
	outputs  terraform.Outputs
	director director.IClient
	db       Opener
	stdout   io.Writer
	stderr   io.Writer
	provider iaas.Provider
}

//NewAWSClient returns a AWS specific implementation of IClient
func NewAWSClient(config config.Config, outputs terraform.Outputs, director director.IClient, stdout, stderr io.Writer, provider iaas.Provider) (IClient, error) {
	directorPublicIP, err := outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	addr := net.JoinHostPort(directorPublicIP, "22")
	key, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
	if err != nil {
		return nil, err
	}
	conf := &ssh.ClientConfig{
		User:            "vcap",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
	}
	var boshDBAddress, boshDBPort string

	boshDBAddress, err = outputs.Get("BoshDBAddress")
	if err != nil {
		return nil, err
	}
	boshDBPort, err = outputs.Get("BoshDBPort")
	if err != nil {
		return nil, err
	}

	db, err := newProxyOpener(addr, conf, &pq.Driver{},
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
			config.RDSUsername,
			config.RDSPassword,
			boshDBAddress,
			boshDBPort,
			config.RDSDefaultDatabaseName,
		),
	)
	if err != nil {
		return nil, err
	}
	return &AWSClient{
		config:   config,
		outputs:  outputs,
		director: director,
		db:       db,
		stdout:   stdout,
		stderr:   stderr,
		provider: provider,
	}, nil
}

//Cleanup is AWS specific implementation of Cleanup
func (client *AWSClient) Cleanup() error {
	return client.director.Cleanup()
}
