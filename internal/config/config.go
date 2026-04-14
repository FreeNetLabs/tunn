package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Auth    Auth   `json:"auth"`
	Local   Local  `json:"local"`
	Payload string `json:"payload"`
	Timeout int    `json:"timeout"`
	TLS     *TLS   `json:"tls,omitempty"`
}

type Auth struct {
	Username string `json:"user"`
	Password string `json:"pass"`
}

type Local struct {
	Port int    `json:"port"`
	Type string `json:"type"`
}

type TLS struct {
	SNI string `json:"sni,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

