package ssh

import (
	"fmt"
	"log"
	"net"
	"strings"
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
			msg := strings.TrimSpace(message)
			if msg != "" {
				log.Printf("ssh banner: %s", msg)
			}
			return nil
		},
	}

	log.Printf("establishing ssh connection for user %s...", c.username)
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
