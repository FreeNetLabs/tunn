package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreeNetLabs/tunn/internal/config"
	"github.com/FreeNetLabs/tunn/internal/connection"
	"github.com/FreeNetLabs/tunn/internal/proxy"
	"github.com/FreeNetLabs/tunn/internal/ssh"
)

type App struct {
	config    *config.Config
	sshClient *ssh.Client
}

func Start(cfg *config.Config) error {
	app := &App{
		config: cfg,
	}
	return app.run()
}

func (app *App) run() error {
	conn, err := connection.Connect(app.config)
	if err != nil {
		return err
	}

	app.sshClient = ssh.NewClient(conn, app.config.Username, app.config.Password)

	if err := app.sshClient.Establish(); err != nil {
		return err
	}

	if err := app.startProxy(); err != nil {
		return err
	}

	fmt.Printf("Tunnel established and %s proxy running on port %d\n", app.config.LocalType, app.config.LocalPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Shutdown signal received, closing tunnel...")

	if app.sshClient != nil {
		app.sshClient.Close()
	}

	fmt.Println("Tunnel closed.")

	return nil
}

func (app *App) startProxy() error {
	switch app.config.LocalType {
	case "socks5", "socks":
		return proxy.StartSOCKS5(app.sshClient, app.config.LocalPort)
	case "http":
		return proxy.StartHTTP(app.sshClient, app.config.LocalPort)
	default:
		return fmt.Errorf("unsupported proxy type: %s", app.config.LocalType)
	}
}
