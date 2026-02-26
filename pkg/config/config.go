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

	config := &Config{}
	content := os.ExpandEnv(string(data))
	if err := json.Unmarshal([]byte(content), config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}
	config.setDefaults()

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

func (c *Config) setDefaults() {
	if c.SSH.Port == 0 {
		c.SSH.Port = 22
	}
	if c.Listener.Port == 0 {
		c.Listener.Port = 1080
	}
	if c.Listener.ProxyType == "" {
		c.Listener.ProxyType = "http"
	}
	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = 30
	}
}
