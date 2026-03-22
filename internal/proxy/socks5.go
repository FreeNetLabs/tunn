package proxy

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/FreeNetLabs/tunn/internal/ssh"
)

type SOCKS5 struct {
	server *Server
}

func NewSOCKS5(sshClient *ssh.Client) *SOCKS5 {
	return &SOCKS5{
		server: NewServer(sshClient),
	}
}

func (s *SOCKS5) Start(localPort int) error {
	return s.server.StartProxy("SOCKS5", localPort, s.handleClient)
}

func (s *SOCKS5) handleClient(clientConn net.Conn) {
	s.server.HandleClientWithTimeout(clientConn, 10*time.Second, func() {
		versionByte := make([]byte, 1)
		if _, err := clientConn.Read(versionByte); err != nil {
			return
		}

		if versionByte[0] == 5 {
			s.handleSOCKS5(clientConn)
		}
	})
}

func (s *SOCKS5) handleSOCKS5(clientConn net.Conn) {
	nmethodsByte := make([]byte, 1)
	if _, err := clientConn.Read(nmethodsByte); err != nil {
		return
	}

	methods := make([]byte, int(nmethodsByte[0]))
	if _, err := io.ReadFull(clientConn, methods); err != nil {
		return
	}

	clientConn.Write([]byte{5, 0})

	requestHeader := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, requestHeader); err != nil {
		return
	}

	cmd := requestHeader[1]
	atyp := requestHeader[3]

	if cmd != 1 {
		s.sendError(clientConn, 7)
		return
	}

	var host string

	switch atyp {
	case 1:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(clientConn, addr); err != nil {
			s.sendError(clientConn, 1)
			return
		}
		host = net.IP(addr).String()

	case 3:
		lengthByte := make([]byte, 1)
		if _, err := clientConn.Read(lengthByte); err != nil {
			s.sendError(clientConn, 1)
			return
		}

		domain := make([]byte, int(lengthByte[0]))
		if _, err := io.ReadFull(clientConn, domain); err != nil {
			s.sendError(clientConn, 1)
			return
		}
		host = string(domain)

	case 4:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(clientConn, addr); err != nil {
			s.sendError(clientConn, 1)
			return
		}
		host = net.IP(addr).String()

	default:
		s.sendError(clientConn, 8)
		return
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, portBytes); err != nil {
		s.sendError(clientConn, 1)
		return
	}
	port := int(binary.BigEndian.Uint16(portBytes))

	s.sendSuccess(clientConn)

	s.server.OpenSSHChannel(clientConn, host, port)
}

func (s *SOCKS5) sendError(clientConn net.Conn, errCode byte) {
	clientConn.Write([]byte{5, errCode, 0, 1, 0, 0, 0, 0, 0, 0})
}

func (s *SOCKS5) sendSuccess(clientConn net.Conn) {
	clientConn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
}
