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

// ProvisionConfig points at the domain provisioning services the portal proxies
// to. Empty URLs mean that provisioning path is disabled.
type ProvisionConfig struct {
	// PerforceURL is the hydraperforceprovision base URL (e.g. over the mesh).
	PerforceURL string `yaml:"perforce_url"`
	// GitURL is the hydragitprovision base URL (e.g. over the mesh).
	GitURL string `yaml:"git_url"`
}

// IAMNimConfig is the identity service the portal sends creators to for sign-in
// and asks for their identity + org memberships.
type IAMNimConfig struct {
	BaseURL string `yaml:"base_url"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Provision ProvisionConfig `yaml:"provision"`
	IAMNim    IAMNimConfig    `yaml:"iamnim"`
}

func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Domain: "hydramancer.experiencenet.com",
		},
		IAMNim: IAMNimConfig{
			BaseURL: "https://iamnim.com",
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
			cfg.applyEnvOverrides()
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	cfg.applyEnvOverrides()
	return cfg, nil
}

// applyEnvOverrides lets deployment set values without a config file. The
// container image carries no config and no state dir, so env is how the scale
// deployment configures the provisioning URL — it survives an image rebuild.
func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("HYDRAMANCER_PROVISION_PERFORCE_URL"); v != "" {
		c.Provision.PerforceURL = v
	}
	if v := os.Getenv("HYDRAMANCER_PROVISION_GIT_URL"); v != "" {
		c.Provision.GitURL = v
	}
	if v := os.Getenv("HYDRAMANCER_IAMNIM_URL"); v != "" {
		c.IAMNim.BaseURL = v
	}
}
