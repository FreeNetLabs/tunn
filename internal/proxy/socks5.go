package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type SOCKS5 struct {
	server *Server
}

func NewSOCKS5(ssh SSHClient) *SOCKS5 {
	return &SOCKS5{
		server: NewServer(ssh),
	}
}

func (s *SOCKS5) Start(localPort int) error {
	return s.server.StartProxy("SOCKS5", localPort, s.handleClient)
}

func (s *SOCKS5) handleClient(clientConn net.Conn) {
	s.server.HandleClientWithTimeout(clientConn, "SOCKS5", 10*time.Second, func() {
		versionByte := make([]byte, 1)
		if _, err := clientConn.Read(versionByte); err != nil {
			fmt.Printf("✗ Error reading SOCKS version: %v\n", err)
			return
		}

		switch versionByte[0] {
		case 5:
			s.handleSOCKS5(clientConn)
		default:
			fmt.Printf("✗ Unsupported SOCKS version: %d (only SOCKS5 supported)\n", versionByte[0])
		}
	})
}

func (s *SOCKS5) handleSOCKS5(clientConn net.Conn) {
	nmethodsByte := make([]byte, 1)
	_, err := clientConn.Read(nmethodsByte)
	if err != nil {
		fmt.Printf("✗ Error reading SOCKS5 nmethods: %v\n", err)
		return
	}

	nmethods := int(nmethodsByte[0])
	methods := make([]byte, nmethods)
	_, err = io.ReadFull(clientConn, methods)
	if err != nil {
		fmt.Printf("✗ Error reading SOCKS5 methods: %v\n", err)
		return
	}

	clientConn.Write([]byte{5, 0})

	requestHeader := make([]byte, 4)
	_, err = io.ReadFull(clientConn, requestHeader)
	if err != nil {
		fmt.Printf("✗ Error reading SOCKS5 request header: %v\n", err)
		return
	}

	cmd := requestHeader[1]
	atyp := requestHeader[3]

	if cmd != 1 {
		s.sendError(clientConn, 7)
		return
	}

	var host string
	var port int

	switch atyp {
	case 1:
		addr := make([]byte, 4)
		_, err = io.ReadFull(clientConn, addr)
		if err != nil {
			s.sendError(clientConn, 1)
			return
		}
		host = fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])

	case 3:
		lengthByte := make([]byte, 1)
		_, err = clientConn.Read(lengthByte)
		if err != nil {
			s.sendError(clientConn, 1)
			return
		}

		length := int(lengthByte[0])
		domain := make([]byte, length)
		_, err = io.ReadFull(clientConn, domain)
		if err != nil {
			s.sendError(clientConn, 1)
			return
		}
		host = string(domain)

	case 4:
		addr := make([]byte, 16)
		_, err = io.ReadFull(clientConn, addr)
		if err != nil {
			s.sendError(clientConn, 1)
			return
		}
		host = fmt.Sprintf("[%x:%x:%x:%x:%x:%x:%x:%x]",
			binary.BigEndian.Uint16(addr[0:2]),
			binary.BigEndian.Uint16(addr[2:4]),
			binary.BigEndian.Uint16(addr[4:6]),
			binary.BigEndian.Uint16(addr[6:8]),
			binary.BigEndian.Uint16(addr[8:10]),
			binary.BigEndian.Uint16(addr[10:12]),
			binary.BigEndian.Uint16(addr[12:14]),
			binary.BigEndian.Uint16(addr[14:16]))

	default:
		s.sendError(clientConn, 8)
		return
	}

	portBytes := make([]byte, 2)
	_, err = io.ReadFull(clientConn, portBytes)
	if err != nil {
		s.sendError(clientConn, 1)
		return
	}
	port = int(binary.BigEndian.Uint16(portBytes))

	s.sendSuccess(clientConn)

	s.server.OpenSSHChannel(clientConn, host, port)
}

func (s *SOCKS5) sendError(clientConn net.Conn, errCode byte) {
	response := []byte{5, errCode, 0, 1, 0, 0, 0, 0, 0, 0}
	clientConn.Write(response)
}

func (s *SOCKS5) sendSuccess(clientConn net.Conn) {
	response := []byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}
	clientConn.Write(response)
}
