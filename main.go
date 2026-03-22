package main

import (
	"flag"
	"log"

	"github.com/FreeNetLabs/tunn/internal/config"
)

func main() {
	configFile := flag.String("config", "config.json", "config file path")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := Start(cfg); err != nil {
		log.Fatalf("failed to start: %v", err)
	}
}
