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
	log.SetFlags(log.Ltime)

	configPath := flag.String("config", "config.json", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config err: %v", err)
	}

	conn, err := transport.Dial(cfg)
	if err != nil {
		log.Fatalf("dial err: %v", err)
	}

	sshClient, err := ssh.NewClient(conn, cfg)
	if err != nil {
		log.Fatalf("ssh err: %v", err)
	}

	switch cfg.Local.Type {
	case "socks":
		err = proxy.ListenAndServeSOCKS5(sshClient, cfg.Local.Port)
	case "http":
		err = proxy.ListenAndServeHTTP(sshClient, cfg.Local.Port)
	default:
		err = fmt.Errorf("unsupported proxy: %s", cfg.Local.Type)
	}

	if err != nil {
		log.Fatalf("proxy err: %v", err)
	}

	log.Printf("%s proxy ready on port %d", cfg.Local.Type, cfg.Local.Port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("closing tunnel...")

	if sshClient != nil {
		sshClient.Close()
	}

	log.Println("closed")
}
