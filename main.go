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

	configFile := flag.String("config", "config.json", "config file")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("config err: %v", err)
	}

	conn, err := transport.Dial(cfg)
	if err != nil {
		log.Fatalf("dial err: %v", err)
	}

	sshClient := ssh.NewClient(conn, cfg.Username, cfg.Password)
	if err := sshClient.Connect(); err != nil {
		log.Fatalf("ssh err: %v", err)
	}

	switch cfg.LocalType {
	case "socks":
		err = proxy.ListenAndServeSOCKS5(sshClient, cfg.LocalPort)
	case "http":
		err = proxy.ListenAndServeHTTP(sshClient, cfg.LocalPort)
	default:
		err = fmt.Errorf("unsupported proxy: %s", cfg.LocalType)
	}

	if err != nil {
		log.Fatalf("proxy err: %v", err)
	}

	log.Printf("%s proxy ready on port %d", cfg.LocalType, cfg.LocalPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("closing tunnel...")

	if sshClient != nil {
		sshClient.Close()
	}

	log.Println("closed")
}
