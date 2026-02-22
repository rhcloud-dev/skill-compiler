package instructions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	return data
}

func TestParseBytes_Valid(t *testing.T) {
	data := readTestdata(t, "valid.md")
	inst, err := ParseBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inst.Frontmatter.Name != "test-tool" {
		t.Errorf("Name = %q, want %q", inst.Frontmatter.Name, "test-tool")
	}
	if inst.Frontmatter.Out != "./output/" {
		t.Errorf("Out = %q, want %q", inst.Frontmatter.Out, "./output/")
	}
	if inst.Frontmatter.Skill.License != "MIT" {
		t.Errorf("Skill.License = %q, want %q", inst.Frontmatter.Skill.License, "MIT")
	}

	// Check sections
	if _, ok := inst.Sections["Product"]; !ok {
		t.Error("missing Product section")
	}
	if _, ok := inst.Sections["Workflows"]; !ok {
		t.Error("missing Workflows section")
	}
	if _, ok := inst.Sections["Examples"]; !ok {
		t.Error("missing Examples section")
	}

	if inst.RawBody == "" {
		t.Error("RawBody is empty")
	}
}

func TestParseBytes_MissingName(t *testing.T) {
	data := readTestdata(t, "missing-name.md")
	_, err := ParseBytes(data)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "missing required field: name") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "missing required field: name")
	}
}

func TestParseBytes_InvalidYAML(t *testing.T) {
	data := readTestdata(t, "invalid-yaml.md")
	_, err := ParseBytes(data)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parsing frontmatter YAML") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "parsing frontmatter YAML")
	}
}

func TestResolveSpecSources_String(t *testing.T) {
	data := []byte("---\nname: test\nspec: ./openapi.yaml\n---\n# Product\nHi")
	inst, err := ParseBytes(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	sources, err := inst.ResolveSpecSources()
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(sources) != 1 || sources[0].Path != "./openapi.yaml" {
		t.Errorf("sources = %+v, want single source with Path=./openapi.yaml", sources)
	}
}

func TestResolveSpecSources_Object(t *testing.T) {
	data := []byte("---\nname: test\nspec:\n  type: cli\n  binary: kubectl\n---\n# Product\nHi")
	inst, err := ParseBytes(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	sources, err := inst.ResolveSpecSources()
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(sources) != 1 || sources[0].Type != "cli" || sources[0].Binary != "kubectl" {
		t.Errorf("sources = %+v, want single CLI source", sources)
	}
}

func TestResolveSpecSources_Array(t *testing.T) {
	data := []byte("---\nname: test\nspec:\n  - ./api.yaml\n  - type: cli\n    binary: mytool\n---\n# Product\nHi")
	inst, err := ParseBytes(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	sources, err := inst.ResolveSpecSources()
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("got %d sources, want 2", len(sources))
	}
	if sources[0].Path != "./api.yaml" {
		t.Errorf("sources[0].Path = %q, want %q", sources[0].Path, "./api.yaml")
	}
	if sources[1].Type != "cli" {
		t.Errorf("sources[1].Type = %q, want %q", sources[1].Type, "cli")
	}
}

func TestValidate_MissingProduct(t *testing.T) {
	data := []byte("---\nname: test\n---\n# Workflows\nSomething")
	inst, err := ParseBytes(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	warnings := inst.Validate()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "missing recommended section: # Product") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about missing Product section, got %v", warnings)
	}
}

func TestEnvPrefix(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"my-app", "MY_APP"},
		{"simple", "SIMPLE"},
		{"multi-word-tool", "MULTI_WORD_TOOL"},
	}
	for _, tt := range tests {
		inst := &Instructions{Frontmatter: Frontmatter{Name: tt.name}}
		got := inst.EnvPrefix()
		if got != tt.want {
			t.Errorf("EnvPrefix(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
