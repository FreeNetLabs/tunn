package connection

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/FreeNetLabs/tunn/pkg/config"
)

type Establisher interface {
	Establish(cfg *config.Config) (net.Conn, error)
}

type DirectEstablisher struct{}

func (d *DirectEstablisher) Establish(cfg *config.Config) (net.Conn, error) {
	sshPort := strconv.Itoa(cfg.SSH.Port)
	address := net.JoinHostPort(cfg.SSH.Host, sshPort)

	fmt.Printf("→ Connecting to %s\n", address)

	var conn net.Conn
	var err error
	if cfg.SSH.Port == 443 {
		tlsConfig := &tls.Config{
			ServerName: cfg.SSH.Host,
			MinVersion: tls.VersionTLS12,
		}
		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: time.Duration(cfg.ConnectionTimeout) * time.Second},
			"tcp",
			address,
			tlsConfig,
		)
	} else {
		conn, err = net.DialTimeout("tcp", address, time.Duration(cfg.ConnectionTimeout)*time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect directly: %w", err)
	}

	if cfg.HTTPPayload != "" {
		wsConn, err := EstablishWSTunnel(conn, cfg.HTTPPayload, cfg.SSH.Host, sshPort, cfg.SSH.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to establish WebSocket tunnel: %w", err)
		}
		return wsConn, nil
	}

	return conn, nil
}

type ProxyEstablisher struct{}

func (p *ProxyEstablisher) Establish(cfg *config.Config) (net.Conn, error) {
	proxyAddress := net.JoinHostPort(cfg.ProxyHost, cfg.ProxyPort)
	sshPort := strconv.Itoa(cfg.SSH.Port)
	fmt.Printf("→ Connecting to proxy %s for target %s\n", proxyAddress, cfg.SSH.Host)

	var conn net.Conn
	var err error
	if cfg.ProxyPort == "443" {
		tlsConfig := &tls.Config{
			ServerName: cfg.ProxyHost,
			MinVersion: tls.VersionTLS12,
		}
		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: time.Duration(cfg.ConnectionTimeout) * time.Second},
			"tcp",
			proxyAddress,
			tlsConfig,
		)
	} else {
		conn, err = net.DialTimeout("tcp", proxyAddress, time.Duration(cfg.ConnectionTimeout)*time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	// Perform WebSocket upgrade through proxy
	wsConn, err := EstablishWSTunnel(conn, cfg.HTTPPayload, cfg.SSH.Host, sshPort, cfg.SSH.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to establish proxy WebSocket tunnel: %w", err)
	}

	fmt.Printf("✓ Proxy WebSocket connection established through %s\n", proxyAddress)
	return wsConn, nil
}

func GetEstablisher(mode string) (Establisher, error) {
	switch mode {
	case "direct":
		return &DirectEstablisher{}, nil
	case "proxy":
		return &ProxyEstablisher{}, nil
	default:
		return nil, fmt.Errorf("unsupported connection mode: %s", mode)
	}
}
