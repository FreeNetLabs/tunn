package ssh

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/html"
)

type Client interface {
	Dial(network, address string) (net.Conn, error)

	Close() error
}

type SSHClient struct {
	conn      net.Conn
	sshClient *ssh.Client
	username  string
	password  string
}

func NewSSHClient(conn net.Conn, username, password string) *SSHClient {
	return &SSHClient{
		conn:     conn,
		username: username,
		password: password,
	}
}

func stripHTMLTags(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return htmlStr
	}
	var b strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return b.String()
}

func (s *SSHClient) StartTransport() error {
	fmt.Println("→ Starting SSH transport over connection...")

	if tcpConn, ok := s.conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	handshakeTimeout := 15 * time.Second
	s.conn.SetDeadline(time.Now().Add(handshakeTimeout))

	config := &ssh.ClientConfig{
		User: s.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         handshakeTimeout,
		BannerCallback: func(message string) error {
			plain := stripHTMLTags(message)
			fmt.Fprintln(os.Stderr, plain)
			return nil
		},
	}

	fmt.Printf("→ Attempting SSH connection with user: %s\n", s.username)

	sshConn, chans, reqs, err := ssh.NewClientConn(s.conn, "tcp", config)
	if err != nil {
		if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
			return fmt.Errorf("SSH handshake timed out after %v", handshakeTimeout)
		}
		return fmt.Errorf("failed to create SSH connection: %v", err)
	}

	s.conn.SetDeadline(time.Time{})

	s.sshClient = ssh.NewClient(sshConn, chans, reqs)
	fmt.Println("✓ SSH transport established and authenticated.")
	return nil
}

func (s *SSHClient) Dial(network, address string) (net.Conn, error) {
	return s.sshClient.Dial(network, address)
}

func (s *SSHClient) Close() error {
	if s.sshClient != nil {
		return s.sshClient.Close()
	}
	return nil
}
