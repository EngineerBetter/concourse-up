package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
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

func writeTempFile(dir, prefix string, data []byte) (name string, err error) {
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return "", err
	}
	name = f.Name()
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(name)
	}
	return name, err
}

// NewRunner returns a new SQL runner
func NewRunner(creds *Credentials) (Runner, error) {
	key, err := ssh.ParsePrivateKey(creds.SSHPrivateKey)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: "vcap",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //TODO: probably don't do this
	}

	var once sync.Once
	var db *sql.DB
	return func(sqlStr string) error {
		var err error
		once.Do(func() {
			var sshClient *ssh.Client
			sshClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", creds.SSHPublicIP), sshConfig)
			if err != nil {
				return
			}

			var caCertFileName string
			caCertFileName, err = writeTempFile("", "concourse-up", []byte(creds.CACert))
			if err != nil {
				return
			}
			defer os.Remove(caCertFileName)
			dialer := &sshDialer{client: sshClient}

			sql.Register("postgres+ssh", dialer)

			db, err = sql.Open("postgres+ssh", postgresArgs(creds, caCertFileName))
			if err != nil {
				return
			}
		})
		if err != nil {
			return err
		}
		rows, err := db.Query(sqlStr)
		if err != nil {
			return err
		}
		if err = rows.Err(); err != nil {
			return err
		}
		return rows.Close()
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
