package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/FreeNetLabs/tunn/internal/ssh"
	"github.com/armon/go-socks5"
)

type SOCKS5 struct {
	sshClient *ssh.Client
}

func NewSOCKS5(sshClient *ssh.Client) *SOCKS5 {
	return &SOCKS5{
		sshClient: sshClient,
	}
}

func (s *SOCKS5) Start(localPort int) error {
	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return s.sshClient.Dial(network, addr)
		},
	}

	server, err := socks5.New(conf)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return err
	}

	go server.Serve(listener)
	return nil
}
