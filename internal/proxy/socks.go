package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func ListenAndServeSOCKS5(dialer Dialer, localPort int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return err
	}

	serve(listener, func(conn net.Conn) {
		handleSOCKS5Connection(conn, dialer)
	})

	return nil
}

func handleSOCKS5Connection(conn net.Conn, dialer Dialer) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	versionByte := make([]byte, 1)
	if _, err := conn.Read(versionByte); err != nil {
		return
	}

	if versionByte[0] != 5 {
		return
	}

	nmethodsByte := make([]byte, 1)
	if _, err := conn.Read(nmethodsByte); err != nil {
		return
	}

	methods := make([]byte, int(nmethodsByte[0]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return
	}

	conn.Write([]byte{5, 0})

	requestHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, requestHeader); err != nil {
		return
	}

	cmd := requestHeader[1]
	atyp := requestHeader[3]

	if cmd != 1 {
		conn.Write([]byte{5, 7, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}

	var host string
	switch atyp {
	case 1:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			conn.Write([]byte{5, 1, 0, 1, 0, 0, 0, 0, 0, 0})
			return
		}
		host = fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])

	case 3:
		lengthByte := make([]byte, 1)
		if _, err := conn.Read(lengthByte); err != nil {
			conn.Write([]byte{5, 1, 0, 1, 0, 0, 0, 0, 0, 0})
			return
		}
		domain := make([]byte, int(lengthByte[0]))
		if _, err := io.ReadFull(conn, domain); err != nil {
			conn.Write([]byte{5, 1, 0, 1, 0, 0, 0, 0, 0, 0})
			return
		}
		host = string(domain)

	case 4:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			conn.Write([]byte{5, 1, 0, 1, 0, 0, 0, 0, 0, 0})
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
		conn.Write([]byte{5, 8, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		conn.Write([]byte{5, 1, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}
	port := int(binary.BigEndian.Uint16(portBytes))

	conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})

	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	address := fmt.Sprintf("%s:%d", host, port)
	remote, err := dialer.Dial("tcp", address)
	if err != nil {
		return
	}
	defer remote.Close()

	relay(conn, remote)
}
