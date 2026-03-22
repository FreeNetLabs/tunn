package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/FreeNetLabs/tunn/internal/ssh"
	"github.com/elazarl/goproxy"
)

type HTTP struct {
	sshClient *ssh.Client
}

func NewHTTP(sshClient *ssh.Client) *HTTP {
	return &HTTP{
		sshClient: sshClient,
	}
}

func (h *HTTP) Start(localPort int) error {
	proxy := goproxy.NewProxyHttpServer()

	proxy.Tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return h.sshClient.Dial(network, addr)
	}
	proxy.ConnectDial = func(network string, addr string) (net.Conn, error) {
		return h.sshClient.Dial(network, addr)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return err
	}

	go http.Serve(listener, proxy)
	return nil
}
