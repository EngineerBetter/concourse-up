package db

import (
	"context"
	"database/sql/driver"
	"io"
	"net"
	"net/url"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

type connectorFunc func(context.Context) (driver.Conn, error)

func (f connectorFunc) Connect(ctx context.Context) (driver.Conn, error) {
	return f(ctx)
}

func (f connectorFunc) Driver() driver.Driver {
	panic("not supported")
}

// NewSSHProxyConnector is used to proxy a database connection through a ssh jumpbox.
// uri is the arguement that will be passed to sqlDriver.Open.
// uri must be a valid URI.
// The return value can be used by sql.OpenDB to obtain a *sql.DB
func NewSSHProxyConnector(jumpboxAddr string, config *ssh.ClientConfig, sqlDriver driver.Driver, uri string) (driver.Connector, error) {
	if jumpboxAddr == "99.99.99.99:22" {
		// Under testing if this branch is taken, skip tunneling
		return connectorFunc(func(_ context.Context) (driver.Conn, error) {
			return sqlDriver.Open(uri)
		}), nil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	remoteHost := u.Host
	client, err := ssh.Dial("tcp", jumpboxAddr, config)
	if err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return // BUG: expose this error somehow, detect Temporary errors.
			}
			go proxyConn(conn, client, remoteHost)
		}
	}()
	u.Host = l.Addr().String()
	newURI := u.String()

	return connectorFunc(func(_ context.Context) (driver.Conn, error) {
		return sqlDriver.Open(newURI)
	}), nil
}

func proxyConn(conn net.Conn, dialer proxy.Dialer, addr string) {
	defer conn.Close()
	target, err := dialer.Dial("tcp", addr)
	if err != nil {
		return //BUG: expose this error somehow
	}
	defer target.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(conn, target)
		wg.Done()
	}()
	go func() {
		io.Copy(target, conn)
		wg.Done()
	}()
	wg.Wait()
}
