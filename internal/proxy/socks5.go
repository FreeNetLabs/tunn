package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/FreeNetLabs/tunn/internal/ssh"
	"github.com/armon/go-socks5"
)

func StartSOCKS5(sshClient *ssh.Client, localPort int) error {
	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return sshClient.Dial(network, addr)
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
