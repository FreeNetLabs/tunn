package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/FreeNetLabs/tunn/internal/tunnel"
	"github.com/FreeNetLabs/tunn/pkg/config"
)

func main() {
	configFile := flag.String("config", "config.json", "config file path")
	flag.StringVar(configFile, "c", "config.json", "config file path (shorthand)")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("Mode: %s\n\n", cfg.Mode)

	manager := tunnel.NewManager(cfg)
	if err := manager.Start(); err != nil {
		log.Fatalf("failed to start tunnel: %v", err)
	}
}
