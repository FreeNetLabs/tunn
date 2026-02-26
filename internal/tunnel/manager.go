package tunnel

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreeNetLabs/tunn/pkg/config"
	"github.com/FreeNetLabs/tunn/pkg/connection"
	"github.com/FreeNetLabs/tunn/pkg/proxy"
	"github.com/FreeNetLabs/tunn/pkg/ssh"
)

type Manager struct {
	config      *config.Config // The tunnel configuration
	sshClient   ssh.Client     // SSH client for tunneling
	proxyServer interface{}    // Local proxy server (SOCKS5 or HTTP)
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

func (m *Manager) Start() error {
	establisher, err := connection.GetEstablisher(m.config.Mode)
	if err != nil {
		return fmt.Errorf("failed to get connection establisher: %w", err)
	}

	conn, err := establisher.Establish(m.config)
	if err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}

	m.sshClient = ssh.NewSSHClient(conn, m.config.SSH.Username, m.config.SSH.Password)

	if sshOverWS, ok := m.sshClient.(*ssh.SSHClient); ok {
		if err := sshOverWS.StartTransport(); err != nil {
			return fmt.Errorf("failed to start SSH transport: %w", err)
		}
	}

	if err := m.startProxy(); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	fmt.Printf("\n✓ Tunnel established and %s proxy running on port %d\n", m.config.Listener.ProxyType, m.config.Listener.Port)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	m.waitForShutdown()

	return nil
}

func (m *Manager) startProxy() error {
	switch m.config.Listener.ProxyType {
	case "socks5", "socks":
		socksProxy := proxy.NewSOCKS5(m.sshClient)
		m.proxyServer = socksProxy
		return socksProxy.Start(m.config.Listener.Port)
	case "http":
		httpProxy := proxy.NewHTTP(m.sshClient)
		m.proxyServer = httpProxy
		return httpProxy.Start(m.config.Listener.Port)
	default:
		return fmt.Errorf("unsupported proxy type: %s", m.config.Listener.ProxyType)
	}
}

func (m *Manager) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n→ Shutdown signal received, closing tunnel...")

	if m.sshClient != nil {
		m.sshClient.Close()
	}

	fmt.Println("✓ Tunnel closed.")
}
