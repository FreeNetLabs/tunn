package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"user"`
	Password  string `json:"pass"`
	LocalPort int    `json:"localPort"`
	LocalType string `json:"localType"`
	Payload   string `json:"payload,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config err: %w", err)
	}

	cfg := &Config{
		Port:      22,
		LocalPort: 1080,
		LocalType: "http",
		Timeout:   30,
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
