package main

import (
	"flag"
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

	sshClient := ssh.NewClient(conn, cfg.Auth.Username, cfg.Auth.Password)
	if err := sshClient.Connect(); err != nil {
		log.Fatalf("ssh err: %v", err)
	}

	err = proxy.ListenAndServe(sshClient, cfg.Local.Type, cfg.Local.Port)
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
