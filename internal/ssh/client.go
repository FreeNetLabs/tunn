package ssh

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	conn     net.Conn
	client   *ssh.Client
	username string
	password string
}

func NewClient(conn net.Conn, username, password string) *Client {
	return &Client{
		conn:     conn,
		username: username,
		password: password,
	}
}

func (c *Client) Connect() error {
	config := &ssh.ClientConfig{
		User: c.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
		BannerCallback: func(message string) error {
			fmt.Print(message)
			return nil
		},
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(c.conn, "tcp", config)
	if err != nil {
		return fmt.Errorf("ssh conn err: %w", err)
	}

	c.client = ssh.NewClient(sshConn, chans, reqs)
	return nil
}

func (c *Client) Dial(network, address string) (net.Conn, error) {
	return c.client.Dial(network, address)
}

func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
