package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/lib/pq"

	"golang.org/x/crypto/ssh"
)

// Credentials represent credentials for connecting to a postgres instance via a jumpbox
type Credentials struct {
	DB            string
	Username      string
	Port          string
	Password      string
	Address       string
	CACert        string
	SSHPrivateKey []byte
	SSHPublicIP   string
}

// Runner is function that runs SQL over a jumpbox
type Runner func(sql string) error

// NewRunner returns a new SQL runner
func NewRunner(creds *Credentials) (Runner, error) {
	key, err := ssh.ParsePrivateKey([]byte(creds.SSHPrivateKey))
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: "vcap",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	return func(sqlStr string) error {
		sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", creds.SSHPublicIP), sshConfig)
		if err != nil {
			return err
		}
		defer sshClient.Close()

		caCertFile, err := ioutil.TempFile("", "concourse-up")
		if err != nil {
			return err
		}
		defer caCertFile.Close()

		if _, err = caCertFile.WriteString(creds.CACert); err != nil {
			return err
		}
		if err = caCertFile.Sync(); err != nil {
			return err
		}

		dialer := &sshDialer{client: sshClient}

		sql.Register("postgres+ssh", dialer)

		db, err := sql.Open("postgres+ssh", postgresArgs(creds, caCertFile.Name()))
		if err != nil {
			return err
		}
		defer db.Close()

		rows, err := db.Query(sqlStr)
		if err != nil {
			return err
		}

		if err = rows.Err(); err != nil {
			return err
		}

		if err = rows.Close(); err != nil {
			return err
		}

		return nil
	}, nil
}

func postgresArgs(creds *Credentials, caCertPath string) string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=verify-full sslrootcert=%s",
		creds.Username,
		creds.Password,
		creds.DB,
		creds.Address,
		creds.Port,
		caCertPath,
	)
}

type sshDialer struct {
	client *ssh.Client
}

func (dialer *sshDialer) Open(s string) (_ driver.Conn, err error) {
	return pq.DialOpen(dialer, s)
}

func (dialer *sshDialer) Dial(network, address string) (net.Conn, error) {
	return dialer.client.Dial(network, address)
}

func (dialer *sshDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return dialer.client.Dial(network, address)
}
