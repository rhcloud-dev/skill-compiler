package codebase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/roberthamel/skill-compiler/internal/instructions"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create a minimal Go project
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.22\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "internal"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "internal", "lib.go"), []byte("package internal\n"), 0o644)

	// Create .gitignore
	_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("dist/\n*.tmp\n"), 0o644)

	// Create files that should be ignored
	_ = os.MkdirAll(filepath.Join(dir, "dist"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "dist", "binary"), []byte("binary"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "test.tmp"), []byte("tmp"), 0o644)

	return dir
}

func TestDetect(t *testing.T) {
	p := New()

	tests := []struct {
		name   string
		source instructions.SpecSource
		want   bool
	}{
		{"codebase type", instructions.SpecSource{Type: "codebase", Path: "."}, true},
		{"codebase no path", instructions.SpecSource{Type: "codebase"}, true},
		{"openapi type", instructions.SpecSource{Type: "openapi"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Detect(tt.source)
			if got != tt.want {
				t.Errorf("Detect(%+v) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestParse_FileTree(t *testing.T) {
	dir := setupTestDir(t)
	p := New()

	source := instructions.SpecSource{Type: "codebase", Path: dir}
	raw, err := p.Fetch(source)
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}

	result, err := p.Parse(raw, source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if result.Structure == nil {
		t.Fatal("Structure should not be nil")
	}

	// Check that gitignored files are excluded
	for _, f := range result.Structure.FileTree {
		if f.Path == "dist/binary" || f.Path == "test.tmp" {
			t.Errorf("gitignored file %q should not be in file tree", f.Path)
		}
	}

	// Check that normal files are included
	found := false
	for _, f := range result.Structure.FileTree {
		if f.Path == "main.go" || filepath.Base(f.Path) == "main.go" {
			found = true
		}
	}
	if !found {
		t.Error("main.go should be in file tree")
	}
}

func TestParse_GoMod(t *testing.T) {
	dir := setupTestDir(t)
	p := New()

	source := instructions.SpecSource{Type: "codebase", Path: dir}
	raw, err := p.Fetch(source)
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}

	result, err := p.Parse(raw, source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if result.Structure == nil || result.Structure.Stack == nil {
		t.Fatal("Stack should not be nil")
	}

	foundGo := false
	for _, lang := range result.Structure.Stack.Languages {
		if lang == "Go" {
			foundGo = true
		}
	}
	if !foundGo {
		t.Errorf("languages = %v, want to contain Go", result.Structure.Stack.Languages)
	}
}

func TestParse_MaxFiles(t *testing.T) {
	dir := t.TempDir()

	// Create many files
	for i := 0; i < 20; i++ {
		_ = os.WriteFile(filepath.Join(dir, "file"+string(rune('a'+i))+".txt"), []byte("content"), 0o644)
	}

	p := New()
	source := instructions.SpecSource{Type: "codebase", Path: dir, MaxFiles: 5}
	raw, err := p.Fetch(source)
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}

	result, err := p.Parse(raw, source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if result.Structure != nil && len(result.Structure.FileTree) > 5 {
		t.Errorf("got %d files, want at most 5 (max-files limit)", len(result.Structure.FileTree))
	}
}
