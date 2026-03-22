package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreeNetLabs/tunn/internal/config"
	"github.com/FreeNetLabs/tunn/internal/connection"
	"github.com/FreeNetLabs/tunn/internal/proxy"
	"github.com/FreeNetLabs/tunn/internal/ssh"
)

type App struct {
	config      *config.Config
	sshClient   ssh.Client
	proxyServer any
}

func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (app *App) Start() error {
	conn, err := connection.EstablishConnection(app.config)
	if err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}

	app.sshClient = ssh.NewSSHClient(conn, app.config.SSH.Username, app.config.SSH.Password)

	if sshOverWS, ok := app.sshClient.(*ssh.SSHClient); ok {
		if err := sshOverWS.StartTransport(); err != nil {
			return fmt.Errorf("failed to start SSH transport: %w", err)
		}
	}

	if err := app.startProxy(); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	fmt.Printf("\n✓ Tunnel established and %s proxy running on port %d\n", app.config.Listener.ProxyType, app.config.Listener.Port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n→ Shutdown signal received, closing tunnel...")

	if app.sshClient != nil {
		app.sshClient.Close()
	}

	fmt.Println("✓ Tunnel closed.")

	return nil
}

func (app *App) startProxy() error {
	switch app.config.Listener.ProxyType {
	case "socks5", "socks":
		socksProxy := proxy.NewSOCKS5(app.sshClient)
		app.proxyServer = socksProxy
		return socksProxy.Start(app.config.Listener.Port)
	case "http":
		httpProxy := proxy.NewHTTP(app.sshClient)
		app.proxyServer = httpProxy
		return httpProxy.Start(app.config.Listener.Port)
	default:
		return fmt.Errorf("unsupported proxy type: %s", app.config.Listener.ProxyType)
	}
}

func main() {
	configFile := flag.String("config", "config.json", "config file path")
	flag.StringVar(configFile, "c", "config.json", "config file path (shorthand)")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app := NewApp(cfg)
	if err := app.Start(); err != nil {
		log.Fatalf("failed to start tunnel: %v", err)
	}
}
