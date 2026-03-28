package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Auth    Auth   `json:"auth"`
	Local   Local  `json:"local"`
	Payload string `json:"payload,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
	TLS     *TLS   `json:"tls,omitempty"`
}

type TLS struct {
	SNI string `json:"sni,omitempty"`
}

type Auth struct {
	Username string `json:"user,omitempty"`
	Password string `json:"pass,omitempty"`
}

type Local struct {
	Port int    `json:"port,omitempty"`
	Type string `json:"type,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config err: %w", err)
	}

	cfg := &Config{
		Port:    80,
		Timeout: 30,
		Local: Local{
			Port: 1080,
			Type: "http",
		},
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
