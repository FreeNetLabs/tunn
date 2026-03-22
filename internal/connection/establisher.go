package connection

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/FreeNetLabs/tunn/internal/config"
)

func EstablishConnection(cfg *config.Config) (net.Conn, error) {
	isProxy := cfg.ProxyHost != "" && cfg.ProxyPort != ""
	sshPortStr := strconv.Itoa(cfg.SSH.Port)

	var dialHost, dialPortStr string

	if isProxy {
		dialHost = cfg.ProxyHost
		dialPortStr = cfg.ProxyPort
		fmt.Printf("→ Connecting to proxy %s:%s for target %s\n", dialHost, dialPortStr, cfg.SSH.Host)
	} else {
		dialHost = cfg.SSH.Host
		dialPortStr = sshPortStr
		fmt.Printf("→ Connecting to %s:%s\n", dialHost, dialPortStr)
	}

	address := net.JoinHostPort(dialHost, dialPortStr)

	conn, err := net.DialTimeout("tcp", address, time.Duration(cfg.ConnectionTimeout)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if cfg.HTTPPayload != "" {
		wsConn, err := EstablishWSTunnel(conn, cfg.HTTPPayload, cfg.SSH.Host, sshPortStr, cfg.SSH.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to establish WebSocket tunnel: %w", err)
		}

		if isProxy {
			fmt.Printf("✓ Proxy WebSocket connection established through %s\n", address)
		}
		return wsConn, nil
	}

	return conn, nil
}
