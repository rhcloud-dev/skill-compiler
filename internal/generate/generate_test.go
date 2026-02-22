package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roberthamel/skill-compiler/internal/instructions"
	"github.com/roberthamel/skill-compiler/internal/ir"
)

func testPipeline(t *testing.T) *Pipeline {
	t.Helper()
	boolTrue := true
	boolFalse := false
	return &Pipeline{
		IR: &ir.IntermediateRepr{},
		Inst: &instructions.Instructions{
			Frontmatter: instructions.Frontmatter{
				Name: "test-tool",
				Out:  "./output/",
				Artifacts: map[string]instructions.Artifact{
					"examples": {Enabled: &boolTrue},
					"scripts":  {Enabled: &boolFalse},
				},
			},
			Sections: map[string]string{
				"Product":         "Product description",
				"Workflows":       "Workflow content",
				"Examples":        "Example content",
				"Common patterns": "Pattern content",
			},
		},
	}
}

func TestEnabledArtifacts_OnlyFilter(t *testing.T) {
	p := testPipeline(t)
	p.Opts.Only = []string{"skill", "llms"}

	artifacts := p.enabledArtifacts()
	if len(artifacts) != 2 {
		t.Fatalf("got %d artifacts, want 2", len(artifacts))
	}
	ids := map[ArtifactID]bool{}
	for _, a := range artifacts {
		ids[a] = true
	}
	if !ids[ArtifactSkill] || !ids[ArtifactLlms] {
		t.Errorf("got %v, want skill and llms", artifacts)
	}
}

func TestEnabledArtifacts_DisabledToggle(t *testing.T) {
	p := testPipeline(t)

	artifacts := p.enabledArtifacts()
	for _, a := range artifacts {
		if a == ArtifactScripts {
			t.Error("scripts should be disabled but was included")
		}
	}
}

func TestArtifactPath_Default(t *testing.T) {
	p := testPipeline(t)

	tests := []struct {
		id   ArtifactID
		want string
	}{
		{ArtifactSkill, filepath.Join("test-tool", "SKILL.md")},
		{ArtifactReference, filepath.Join("test-tool", "references", "reference.md")},
		{ArtifactLlms, "llms.txt"},
		{ArtifactChangelog, "CHANGELOG.md"},
	}
	for _, tt := range tests {
		got := p.artifactPath(tt.id)
		if got != tt.want {
			t.Errorf("artifactPath(%s) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestArtifactPath_CustomFilename(t *testing.T) {
	p := testPipeline(t)
	p.Inst.Frontmatter.Artifacts["skill"] = instructions.Artifact{Filename: "custom.md"}

	got := p.artifactPath(ArtifactSkill)
	want := filepath.Join("test-tool", "custom.md")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRelevantSections(t *testing.T) {
	p := testPipeline(t)

	// Examples should only include Workflows, Examples, Common patterns
	sections := p.RelevantSections(ArtifactExamples)
	if !strings.Contains(sections, "Workflows") {
		t.Error("examples sections should contain Workflows")
	}
	if !strings.Contains(sections, "Examples") {
		t.Error("examples sections should contain Examples")
	}
	if strings.Contains(sections, "Product") {
		t.Error("examples sections should not contain Product")
	}

	// llms should only include Product
	sections = p.RelevantSections(ArtifactLlms)
	if !strings.Contains(sections, "Product") {
		t.Error("llms sections should contain Product")
	}
	if strings.Contains(sections, "Workflows") {
		t.Error("llms sections should not contain Workflows")
	}

	// reference should have no specific sections
	sections = p.RelevantSections(ArtifactReference)
	if sections != "" {
		t.Errorf("reference sections should be empty, got %q", sections)
	}
}

func TestPrependChangelogEntry_New(t *testing.T) {
	result := PrependChangelogEntry("### Added\n- Feature X", "")
	if !strings.HasPrefix(result, "# CHANGELOG") {
		t.Error("new changelog should start with # CHANGELOG header")
	}
	if !strings.Contains(result, "### Added") {
		t.Error("should contain the entry content")
	}
	if !strings.Contains(result, "—") {
		t.Error("date header should contain em-dash")
	}
}

func TestPrependChangelogEntry_Existing(t *testing.T) {
	existing := "# CHANGELOG\n\n## 2025-01-01 — Wednesday\n\n### Added\n- Old feature"
	result := PrependChangelogEntry("### Added\n- New feature", existing)

	if !strings.HasPrefix(result, "# CHANGELOG") {
		t.Error("should preserve header")
	}
	// New entry should come before old
	newIdx := strings.Index(result, "New feature")
	oldIdx := strings.Index(result, "Old feature")
	if newIdx > oldIdx {
		t.Error("new entry should be before old entry")
	}
}

func TestWriteScripts(t *testing.T) {
	dir := t.TempDir()
	content := "```health-check.sh\n#!/bin/bash\necho \"OK\"\n```\n\n```discover.sh\n#!/bin/bash\nls\n```"

	if err := writeScripts(dir, "scripts", content); err != nil {
		t.Fatalf("writeScripts error: %v", err)
	}

	// Check first script
	data, err := os.ReadFile(filepath.Join(dir, "scripts", "health-check.sh"))
	if err != nil {
		t.Fatalf("reading health-check.sh: %v", err)
	}
	if !strings.Contains(string(data), "echo \"OK\"") {
		t.Errorf("health-check.sh = %q, want to contain echo OK", string(data))
	}

	// Check second script
	data, err = os.ReadFile(filepath.Join(dir, "scripts", "discover.sh"))
	if err != nil {
		t.Fatalf("reading discover.sh: %v", err)
	}
	if !strings.Contains(string(data), "ls") {
		t.Errorf("discover.sh = %q, want to contain ls", string(data))
	}
}

func TestUserMessage_Changelog(t *testing.T) {
	p := testPipeline(t)
	p.Opts.PrevArtifacts = map[ArtifactID]string{
		ArtifactSkill:     "previous skill content",
		ArtifactReference: "previous reference content",
		ArtifactChangelog: "previous changelog content",
	}

	msg := p.userMessage(ArtifactChangelog)
	if !strings.Contains(msg, "Previous skill") {
		t.Error("changelog message should include previous skill")
	}
	if !strings.Contains(msg, "Previous reference") {
		t.Error("changelog message should include previous reference")
	}
	if !strings.Contains(msg, "Previous CHANGELOG.md") {
		t.Error("changelog message should include previous changelog")
	}
}

func TestUserMessage_FirstGeneration(t *testing.T) {
	p := testPipeline(t)
	p.Opts.PrevArtifacts = map[ArtifactID]string{}

	msg := p.userMessage(ArtifactChangelog)
	if !strings.Contains(msg, "first generation") {
		t.Error("first-gen changelog should note no previous artifacts")
	}
}
