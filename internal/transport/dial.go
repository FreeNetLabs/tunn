package transport

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/FreeNetLabs/tunn/internal/config"
)

func Dial(cfg *config.Config) (net.Conn, error) {
	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	log.Printf("connecting to %s...", address)
	conn, err := net.DialTimeout("tcp", address, time.Duration(cfg.Timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	log.Printf("connected to %s", address)

	if cfg.Payload != "" {
		if err := injectPayload(conn, cfg.Payload); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func injectPayload(conn net.Conn, payload string) error {
	log.Println("injecting payload...")
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
		return fmt.Errorf("ws upgrade failed: %s", string(data))
	}

	firstLine := string(bytes.SplitN(bytes.TrimSpace(data), []byte("\r\n"), 2)[0])
	log.Printf("server responded: %s", firstLine)
	return nil
}
