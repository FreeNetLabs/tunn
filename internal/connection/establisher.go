package connection

import (
	"net"
	"strconv"
	"time"

	"github.com/FreeNetLabs/tunn/internal/config"
)

func Connect(cfg *config.Config) (net.Conn, error) {
	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	conn, err := net.DialTimeout("tcp", address, time.Duration(cfg.Timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	if cfg.Payload != "" {
		if err := injectPayload(conn, cfg.Payload); err != nil {
			return nil, err
		}
	}

	return conn, nil
}
