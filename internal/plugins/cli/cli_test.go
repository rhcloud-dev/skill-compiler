package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roberthamel/skill-compiler/internal/instructions"
)

func TestDetect(t *testing.T) {
	p := New()

	tests := []struct {
		name   string
		source instructions.SpecSource
		want   bool
	}{
		{"cli with binary", instructions.SpecSource{Type: "cli", Binary: "ls"}, true},
		{"cli no binary", instructions.SpecSource{Type: "cli"}, false},
		{"openapi type", instructions.SpecSource{Type: "openapi", Path: "api.yaml"}, false},
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

func TestParseHelpOutput(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "standard-help.txt"))
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}
	result := parseHelpOutput(string(data))

	// Should extract description
	if result.description == "" {
		t.Error("description should not be empty")
	}

	// Should find subcommands
	if len(result.subcommands) != 3 {
		t.Errorf("got %d subcommands, want 3, got: %v", len(result.subcommands), result.subcommands)
	}
	subSet := map[string]bool{}
	for _, s := range result.subcommands {
		subSet[s] = true
	}
	for _, want := range []string{"serve", "config", "version"} {
		if !subSet[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}

	// Should find flags
	if len(result.flags) < 2 {
		t.Errorf("got %d flags, want at least 2", len(result.flags))
	}
	flagNames := map[string]bool{}
	for _, f := range result.flags {
		flagNames[f.name] = true
	}
	if !flagNames["--help"] {
		t.Error("missing --help flag")
	}
	if !flagNames["--verbose"] {
		t.Error("missing --verbose flag")
	}

	// Should find aliases
	if len(result.aliases) == 0 {
		t.Error("should have extracted aliases")
	}
}

func TestSplitCommandBlocks(t *testing.T) {
	input := "=== COMMAND: mytool ===\nUsage: mytool [cmd]\n=== END ===\n\n=== COMMAND: mytool serve ===\nStart server\n=== END ==="

	blocks := splitCommandBlocks(input)
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].command != "mytool" {
		t.Errorf("blocks[0].command = %q, want %q", blocks[0].command, "mytool")
	}
	if blocks[1].command != "mytool serve" {
		t.Errorf("blocks[1].command = %q, want %q", blocks[1].command, "mytool serve")
	}
}

func TestParse_CommandBlocks(t *testing.T) {
	p := New()
	input := "=== COMMAND: mytool ===\nmytool — a tool\n\nCommands:\n  serve         Start server\n  config        Manage config\n\nFlags:\n  -v, --verbose   bool   Enable verbose\n=== END ==="

	source := instructions.SpecSource{Type: "cli", Binary: "mytool"}
	result, err := p.Parse([]byte(input), source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(result.Operations) != 1 {
		t.Fatalf("got %d operations, want 1", len(result.Operations))
	}

	op := result.Operations[0]
	if !strings.Contains(op.RawHelpText, "mytool") {
		t.Error("raw help text should be preserved")
	}
	if len(op.Parameters) == 0 {
		t.Error("should have parsed flags into parameters")
	}
}
