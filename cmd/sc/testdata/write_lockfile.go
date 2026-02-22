//go:build ignore

// write_lockfile is a test helper that computes cache hashes for all artifacts
// and writes a .sc-lock.json file. It uses the same logic as sc diff to ensure
// hash compatibility.
//
// Usage: go run ./cmd/sc/testdata/write_lockfile.go <target-dir>
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/roberthamel/skill-compiler/internal/cache"
	"github.com/roberthamel/skill-compiler/internal/generate"
	"github.com/roberthamel/skill-compiler/internal/instructions"
	"github.com/roberthamel/skill-compiler/internal/ir"
	cliplugin "github.com/roberthamel/skill-compiler/internal/plugins/cli"
	"github.com/roberthamel/skill-compiler/internal/plugins/codebase"
	"github.com/roberthamel/skill-compiler/internal/plugins/openapi"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: write_lockfile <target-dir>")
		os.Exit(1)
	}
	targetDir := os.Args[1]

	// Change to target dir so relative paths in instructions resolve
	if err := os.Chdir(targetDir); err != nil {
		fmt.Fprintf(os.Stderr, "chdir: %v\n", err)
		os.Exit(1)
	}

	inst, err := instructions.Parse("COMPILER_INSTRUCTIONS.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		os.Exit(1)
	}

	sources, err := inst.ResolveSpecSources()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve: %v\n", err)
		os.Exit(1)
	}

	reg := ir.NewRegistry()
	reg.Register(openapi.New())
	reg.Register(cliplugin.New())
	reg.Register(codebase.New())

	parsedIR, _, err := reg.ProcessSources(sources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "process: %v\n", err)
		os.Exit(1)
	}

	irJSON, _ := json.Marshal(parsedIR)
	specContent := string(irJSON)

	pipeline := &generate.Pipeline{IR: parsedIR, Inst: inst}

	lf := &cache.LockFile{Artifacts: make(map[string]cache.LockEntry)}
	for _, id := range generate.AllArtifacts {
		prompt := pipeline.SystemPromptFor(id)
		sections := pipeline.RelevantSections(id)
		inputHash := cache.HashInput(specContent, sections, prompt)
		lf.Artifacts[string(id)] = cache.LockEntry{
			InputHash:  inputHash,
			OutputHash: "placeholder",
			Timestamp:  "2025-01-01T00:00:00Z",
		}
	}

	if err := cache.SaveLockFile(targetDir, lf); err != nil {
		fmt.Fprintf(os.Stderr, "save: %v\n", err)
		os.Exit(1)
	}
}
