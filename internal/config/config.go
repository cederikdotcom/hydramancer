package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Domain string `yaml:"domain"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
}

func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Domain: "hydramancer.experiencenet.com",
		},
	}
}

func (c Config) GetCertCache() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hydramancer", "certs")
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, ".hydramancer", "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
