package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Mode      string `json:"mode"`
	ProxyHost string `json:"proxyHost,omitempty"`
	ProxyPort string `json:"proxyPort,omitempty"`

	SSH SSHConfig `json:"ssh"`

	Listener ListenerConfig `json:"listener"`

	HTTPPayload       string `json:"httpPayload,omitempty"`
	ConnectionTimeout int    `json:"connectionTimeout,omitempty"`
}

type SSHConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ListenerConfig struct {
	Port      int    `json:"port"`
	ProxyType string `json:"proxyType"`
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("no config file specified")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{
		SSH: SSHConfig{
			Port: 22,
		},
		Listener: ListenerConfig{
			Port:      1080,
			ProxyType: "http",
		},
		ConnectionTimeout: 30,
	}
	content := os.ExpandEnv(string(data))
	if err := json.Unmarshal([]byte(content), config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	validModes := map[string]bool{"direct": true, "proxy": true}
	if !validModes[c.Mode] {
		return fmt.Errorf("invalid mode '%s', must be one of: direct, proxy", c.Mode)
	}

	if c.SSH.Host == "" {
		return fmt.Errorf("SSH host is required")
	}
	if c.SSH.Username == "" {
		return fmt.Errorf("SSH username is required")
	}
	if c.SSH.Password == "" {
		return fmt.Errorf("SSH password is required")
	}

	if c.Mode == "proxy" {
		if c.ProxyHost == "" || c.ProxyPort == "" {
			return fmt.Errorf("proxyHost and proxyPort are required for proxy mode")
		}
	}

	return nil
}
