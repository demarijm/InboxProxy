package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Addr       string `json:"addr"`
	LogLevel   string `json:"log_level"`
	ConfigFile string `json:"-"`
}

func Default() *Config {
	return &Config{
		Addr:     ":8080",
		LogLevel: "info",
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
		cfg.ConfigFile = path
	}

	if v := os.Getenv("ADDR"); v != "" {
		cfg.Addr = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("addr required")
	}
	return nil
}

func (c *Config) Reload() error {
	if c.ConfigFile == "" {
		return nil
	}
	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return err
	}
	var updated Config
	if err := json.Unmarshal(data, &updated); err != nil {
		return err
	}
	if err := updated.Validate(); err != nil {
		return err
	}
	updated.ConfigFile = c.ConfigFile
	*c = updated
	if v := os.Getenv("ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	return nil
}
