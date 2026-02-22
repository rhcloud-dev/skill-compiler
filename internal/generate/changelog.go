package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LoadPreviousArtifacts reads existing artifacts from the output directory.
func LoadPreviousArtifacts(outputDir, skillName string) map[ArtifactID]string {
	prev := make(map[ArtifactID]string)

	paths := map[ArtifactID]string{
		ArtifactSkill:     filepath.Join(outputDir, skillName, "SKILL.md"),
		ArtifactReference: filepath.Join(outputDir, skillName, "references", "reference.md"),
		ArtifactExamples:  filepath.Join(outputDir, skillName, "references", "examples.md"),
		ArtifactLlms:      filepath.Join(outputDir, "llms.txt"),
		ArtifactLlmsAPI:   filepath.Join(outputDir, "llms-api.txt"),
		ArtifactLlmsFull:  filepath.Join(outputDir, "llms-full.txt"),
		ArtifactChangelog: filepath.Join(outputDir, "CHANGELOG.md"),
	}

	for id, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			prev[id] = string(data)
		}
	}

	return prev
}

// PrependChangelogEntry prepends a new entry to an existing CHANGELOG.md,
// preserving previous entries.
func PrependChangelogEntry(newEntry, existingChangelog string) string {
	now := time.Now()
	date := now.Format("2006-01-02")
	weekday := now.Format("Monday")
	header := fmt.Sprintf("## %s — %s\n\n", date, weekday)

	entry := header + strings.TrimSpace(newEntry) + "\n"

	if existingChangelog == "" {
		return "# CHANGELOG\n\n" + entry
	}

	// Find where previous entries start (after the # CHANGELOG header)
	lines := strings.SplitN(existingChangelog, "\n", 3)
	if len(lines) >= 1 && strings.HasPrefix(strings.TrimSpace(lines[0]), "# ") {
		// Has a top-level header — insert after it
		rest := ""
		if len(lines) >= 3 {
			rest = lines[2]
		}
		return lines[0] + "\n\n" + entry + "\n" + rest
	}

	// No header — just prepend
	return entry + "\n" + existingChangelog
}
