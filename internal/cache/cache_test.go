package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashInput_Deterministic(t *testing.T) {
	h1 := HashInput("spec", "instructions", "prompt")
	h2 := HashInput("spec", "instructions", "prompt")
	if h1 != h2 {
		t.Errorf("hashes differ for same input: %s vs %s", h1, h2)
	}
}

func TestHashInput_Sensitive(t *testing.T) {
	h1 := HashInput("spec-a", "instructions", "prompt")
	h2 := HashInput("spec-b", "instructions", "prompt")
	if h1 == h2 {
		t.Error("hashes should differ for different input")
	}
}

func TestHashOutput_Deterministic(t *testing.T) {
	h1 := HashOutput("content")
	h2 := HashOutput("content")
	if h1 != h2 {
		t.Errorf("hashes differ for same output: %s vs %s", h1, h2)
	}
}

func TestLockFile_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	lf := &LockFile{Artifacts: map[string]LockEntry{
		"skill": {InputHash: "abc", OutputHash: "def", Model: "test-model"},
	}}

	if err := SaveLockFile(dir, lf); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadLockFile(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	entry, ok := loaded.Artifacts["skill"]
	if !ok {
		t.Fatal("missing skill entry")
	}
	if entry.InputHash != "abc" || entry.OutputHash != "def" || entry.Model != "test-model" {
		t.Errorf("entry = %+v, want abc/def/test-model", entry)
	}
}

func TestIsUpToDate(t *testing.T) {
	lf := &LockFile{Artifacts: map[string]LockEntry{
		"skill": {InputHash: "abc123"},
	}}

	if !lf.IsUpToDate("skill", "abc123") {
		t.Error("should be up to date with matching hash")
	}
	if lf.IsUpToDate("skill", "different") {
		t.Error("should not be up to date with different hash")
	}
	if lf.IsUpToDate("nonexistent", "abc123") {
		t.Error("should not be up to date for missing artifact")
	}
}

func TestCachedReadWrite(t *testing.T) {
	dir := t.TempDir()
	content := "cached artifact content"

	if err := WriteCached(dir, "skill", content); err != nil {
		t.Fatalf("write error: %v", err)
	}

	got, err := ReadCached(dir, "skill")
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}

	// Verify file exists in .sc-cache/
	path := filepath.Join(dir, ".sc-cache", "skill")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("cache file not found: %v", err)
	}
}
