package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreeNetLabs/tunn/internal/config"
	"github.com/FreeNetLabs/tunn/internal/proxy"
	"github.com/FreeNetLabs/tunn/internal/ssh"
	"github.com/FreeNetLabs/tunn/internal/transport"
)

func main() {
	configFile := flag.String("config", "config.json", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	conn, err := transport.Dial(cfg)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	sshClient := ssh.NewClient(conn, cfg.Username, cfg.Password)
	if err := sshClient.Connect(); err != nil {
		log.Fatalf("failed to establish ssh tunnel: %v", err)
	}

	switch cfg.LocalType {
	case "socks":
		err = proxy.ListenAndServeSOCKS5(sshClient, cfg.LocalPort)
	case "http":
		err = proxy.ListenAndServeHTTP(sshClient, cfg.LocalPort)
	default:
		err = fmt.Errorf("unsupported proxy type: %s", cfg.LocalType)
	}

	if err != nil {
		log.Fatalf("failed to start proxy: %v", err)
	}

	fmt.Printf("Tunnel established and %s proxy running on port %d\n", cfg.LocalType, cfg.LocalPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Shutdown signal received, closing tunnel...")

	if sshClient != nil {
		sshClient.Close()
	}

	fmt.Println("Tunnel closed.")
}
