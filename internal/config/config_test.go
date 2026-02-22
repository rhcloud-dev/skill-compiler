package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTempConfig overrides the config dir for testing.
func setupTempConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".config", "sc")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Override HOME so configDir() points to temp
	t.Setenv("HOME", dir)
	// Clear any SC_ env vars that might interfere
	t.Setenv("SC_PROVIDER", "")
	t.Setenv("SC_API_KEY", "")
	t.Setenv("SC_MODEL", "")
	t.Setenv("SC_BASE_URL", "")
	return dir
}

func TestSetAndLoad(t *testing.T) {
	setupTempConfig(t)

	if err := Set("provider", "anthropic"); err != nil {
		t.Fatalf("set error: %v", err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if cfg.Provider != "anthropic" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "anthropic")
	}
}

func TestList_MasksAPIKey(t *testing.T) {
	setupTempConfig(t)

	if err := Set("api-key", "sk-1234567890abcdef"); err != nil {
		t.Fatalf("set error: %v", err)
	}
	m, err := List()
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	key := m["api-key"]
	if key == "sk-1234567890abcdef" {
		t.Error("API key should be masked")
	}
	if !strings.HasPrefix(key, "sk-1") {
		t.Errorf("masked key should start with first 4 chars, got %q", key)
	}
	if !strings.HasSuffix(key, "cdef") {
		t.Errorf("masked key should end with last 4 chars, got %q", key)
	}
}

func TestReset(t *testing.T) {
	setupTempConfig(t)

	if err := Set("provider", "openai"); err != nil {
		t.Fatalf("set error: %v", err)
	}
	if err := Reset(); err != nil {
		t.Fatalf("reset error: %v", err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if cfg.Provider != "" {
		t.Errorf("Provider = %q after reset, want empty", cfg.Provider)
	}
}

func TestSet_UnknownKey(t *testing.T) {
	setupTempConfig(t)

	err := Set("unknown-key", "value")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "unknown config key")
	}
}

func TestResolve_Priority(t *testing.T) {
	setupTempConfig(t)

	// Set config file value
	if err := Set("provider", "from-config"); err != nil {
		t.Fatal(err)
	}

	// Set env var (higher priority)
	t.Setenv("SC_PROVIDER", "from-env")

	// Frontmatter (higher than env)
	fm := &Config{Provider: "from-frontmatter"}

	// CLI flag (highest)
	resolved, err := Resolve("from-cli", "", "", "", fm)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if resolved.Provider != "from-cli" {
		t.Errorf("Provider = %q, want %q (CLI flag should win)", resolved.Provider, "from-cli")
	}

	// Without CLI flag, frontmatter wins
	resolved, err = Resolve("", "", "", "", fm)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if resolved.Provider != "from-frontmatter" {
		t.Errorf("Provider = %q, want %q (frontmatter should win over env)", resolved.Provider, "from-frontmatter")
	}

	// Without CLI and frontmatter, env wins
	resolved, err = Resolve("", "", "", "", nil)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if resolved.Provider != "from-env" {
		t.Errorf("Provider = %q, want %q (env should win over config)", resolved.Provider, "from-env")
	}
}
