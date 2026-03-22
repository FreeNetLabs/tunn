package proxy

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/FreeNetLabs/tunn/internal/ssh"
)

type Server struct {
	ssh *ssh.Client
}

func NewServer(sshClient *ssh.Client) *Server {
	return &Server{ssh: sshClient}
}

func (s *Server) StartProxy(proxyType string, localPort int, handler func(net.Conn)) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return err
	}

	go func() {
		defer listener.Close()
		for {
			clientConn, err := listener.Accept()
			if err == nil {
				go handler(clientConn)
			}
		}
	}()

	return nil
}

func (s *Server) HandleClientWithTimeout(clientConn net.Conn, timeout time.Duration, handler func()) {
	defer clientConn.Close()
	clientConn.SetDeadline(time.Now().Add(timeout))
	handler()
}

func (s *Server) OpenSSHChannel(clientConn net.Conn, host string, port int) {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	sshConn, err := s.ssh.Dial("tcp", address)
	if err != nil {
		return
	}
	defer sshConn.Close()

	s.forwardData(clientConn, sshConn)
}

func (s *Server) forwardData(conn1, conn2 net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(conn1, conn2)
	}()

	go func() {
		defer wg.Done()
		io.Copy(conn2, conn1)
	}()

	wg.Wait()
}
