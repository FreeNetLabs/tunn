package connection

import (
	"bytes"
	"fmt"
	"net"
)

func injectPayload(conn net.Conn, payload string) error {
	if _, err := conn.Write([]byte(payload)); err != nil {
		conn.Close()
		return err
	}

	var data []byte
	buf := make([]byte, 1)
	for {
		if _, err := conn.Read(buf); err != nil {
			conn.Close()
			return err
		}
		data = append(data, buf[0])
		if bytes.HasSuffix(data, []byte("\r\n\r\n")) {
			break
		}
	}

	if !bytes.Contains(data, []byte("101")) {
		conn.Close()
		return fmt.Errorf("websocket upgrade failed")
	}

	return nil
}
