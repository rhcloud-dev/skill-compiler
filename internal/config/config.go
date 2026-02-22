package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the CLI configuration values.
type Config struct {
	Provider string `yaml:"provider,omitempty" mapstructure:"provider"`
	APIKey   string `yaml:"api-key,omitempty" mapstructure:"api-key"`
	Model    string `yaml:"model,omitempty" mapstructure:"model"`
	BaseURL  string `yaml:"base-url,omitempty" mapstructure:"base-url"`
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

// newViper creates a configured viper instance for sc config.
func newViper() (*viper.Viper, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	// Bind SC_* env vars
	v.SetEnvPrefix("SC")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Read config file (ignore not-found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only ignore "not found" — other errors (parse, permission) bubble up
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("reading config: %w", err)
			}
		}
	}

	return v, nil
}

// Load reads the config file from ~/.config/sc/config.yaml.
// Returns an empty Config if the file doesn't exist.
func Load() (*Config, error) {
	v, err := newViper()
	if err != nil {
		return nil, err
	}
	return &Config{
		Provider: v.GetString("provider"),
		APIKey:   v.GetString("api-key"),
		Model:    v.GetString("model"),
		BaseURL:  v.GetString("base-url"),
	}, nil
}

// Set updates a single key in the config file.
func Set(key, value string) error {
	if !isValidKey(key) {
		return fmt.Errorf("unknown config key %q (valid keys: %s)", key, strings.Join(ValidKeys, ", "))
	}

	v, err := newViper()
	if err != nil {
		return err
	}

	v.Set(key, value)

	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	configFile := filepath.Join(dir, "config.yaml")
	v.SetConfigFile(configFile)
	return v.WriteConfig()
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
	dir, err := configDir()
	if err != nil {
		return err
	}
	p := filepath.Join(dir, "config.yaml")
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

func isValidKey(key string) bool {
	for _, k := range ValidKeys {
		if k == key {
			return true
		}
	}
	return false
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
// Viper handles config file + env vars automatically. We layer
// frontmatter and CLI flags on top.
func Resolve(cliProvider, cliModel, cliAPIKey, cliBaseURL string, frontmatter *Config) (*Resolved, error) {
	v, err := newViper()
	if err != nil {
		return nil, err
	}

	// Viper already merged: config file < env vars (SC_PROVIDER, SC_API_KEY, etc.)
	r := &Resolved{
		Provider: v.GetString("provider"),
		APIKey:   v.GetString("api-key"),
		Model:    v.GetString("model"),
		BaseURL:  v.GetString("base-url"),
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
