package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the CLI configuration values.
type Config struct {
	Provider string `yaml:"provider,omitempty"`
	APIKey   string `yaml:"api-key,omitempty"`
	Model    string `yaml:"model,omitempty"`
	BaseURL  string `yaml:"base-url,omitempty"`
}

// ValidKeys lists the allowed config keys.
var ValidKeys = []string{"provider", "api-key", "model", "base-url"}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, ".config", "sc"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file from ~/.config/sc/config.yaml.
// Returns an empty Config if the file doesn't exist.
func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to ~/.config/sc/config.yaml.
func Save(cfg *Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(p, data, 0o644)
}

// Set updates a single key in the config.
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	switch key {
	case "provider":
		cfg.Provider = value
	case "api-key":
		cfg.APIKey = value
	case "model":
		cfg.Model = value
	case "base-url":
		cfg.BaseURL = value
	default:
		return fmt.Errorf("unknown config key %q (valid keys: %s)", key, strings.Join(ValidKeys, ", "))
	}
	return Save(cfg)
}

// List returns key-value pairs for display, masking the API key.
func List() (map[string]string, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	m := map[string]string{
		"provider": cfg.Provider,
		"api-key":  maskKey(cfg.APIKey),
		"model":    cfg.Model,
		"base-url": cfg.BaseURL,
	}
	return m, nil
}

// Reset removes the config file.
func Reset() error {
	p, err := configPath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing config: %w", err)
	}
	return nil
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

// Resolved holds the final resolved provider settings after merging all sources.
type Resolved struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

// Resolve merges provider settings in priority order:
// CLI flags > frontmatter > env vars > config file.
func Resolve(cliProvider, cliModel, cliAPIKey, cliBaseURL string, frontmatter *Config) (*Resolved, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	r := &Resolved{
		Provider: cfg.Provider,
		APIKey:   cfg.APIKey,
		Model:    cfg.Model,
		BaseURL:  cfg.BaseURL,
	}

	// Env vars override config file
	if v := os.Getenv("SC_PROVIDER"); v != "" {
		r.Provider = v
	}
	if v := os.Getenv("SC_API_KEY"); v != "" {
		r.APIKey = v
	}
	if v := os.Getenv("SC_MODEL"); v != "" {
		r.Model = v
	}
	if v := os.Getenv("SC_BASE_URL"); v != "" {
		r.BaseURL = v
	}

	// Frontmatter overrides env vars
	if frontmatter != nil {
		if frontmatter.Provider != "" {
			r.Provider = frontmatter.Provider
		}
		if frontmatter.APIKey != "" {
			r.APIKey = frontmatter.APIKey
		}
		if frontmatter.Model != "" {
			r.Model = frontmatter.Model
		}
		if frontmatter.BaseURL != "" {
			r.BaseURL = frontmatter.BaseURL
		}
	}

	// CLI flags override frontmatter
	if cliProvider != "" {
		r.Provider = cliProvider
	}
	if cliAPIKey != "" {
		r.APIKey = cliAPIKey
	}
	if cliModel != "" {
		r.Model = cliModel
	}
	if cliBaseURL != "" {
		r.BaseURL = cliBaseURL
	}

	// Also check provider-specific env vars as fallback for API key
	if r.APIKey == "" {
		switch strings.ToLower(r.Provider) {
		case "anthropic":
			r.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		case "openai":
			r.APIKey = os.Getenv("OPENAI_API_KEY")
		}
	}

	return r, nil
}
