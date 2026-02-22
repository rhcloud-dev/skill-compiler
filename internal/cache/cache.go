package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LockFile represents the .sc-lock.json structure.
type LockFile struct {
	Artifacts map[string]LockEntry `json:"artifacts"`
}

// LockEntry records hashes and metadata for a single artifact.
type LockEntry struct {
	InputHash  string `json:"inputHash"`
	OutputHash string `json:"outputHash"`
	Timestamp  string `json:"timestamp"`
	Model      string `json:"model"`
}

// HashInput computes a SHA-256 hash of the given inputs for an artifact.
func HashInput(specContent, instructionsSections, systemPrompt string) string {
	h := sha256.New()
	h.Write([]byte(specContent))
	h.Write([]byte(instructionsSections))
	h.Write([]byte(systemPrompt))
	return hex.EncodeToString(h.Sum(nil))
}

// HashOutput computes a SHA-256 hash of the artifact output.
func HashOutput(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// LoadLockFile reads .sc-lock.json from the project directory.
func LoadLockFile(dir string) (*LockFile, error) {
	path := filepath.Join(dir, ".sc-lock.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{Artifacts: make(map[string]LockEntry)}, nil
		}
		return nil, fmt.Errorf("reading lockfile: %w", err)
	}
	var lf LockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parsing lockfile: %w", err)
	}
	if lf.Artifacts == nil {
		lf.Artifacts = make(map[string]LockEntry)
	}
	return &lf, nil
}

// SaveLockFile writes .sc-lock.json to the project directory.
func SaveLockFile(dir string, lf *LockFile) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling lockfile: %w", err)
	}
	path := filepath.Join(dir, ".sc-lock.json")
	return os.WriteFile(path, data, 0o644)
}

// UpdateEntry updates a single artifact entry in the lockfile.
func (lf *LockFile) UpdateEntry(artifactID, inputHash, outputHash, model string) {
	lf.Artifacts[artifactID] = LockEntry{
		InputHash:  inputHash,
		OutputHash: outputHash,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Model:      model,
	}
}

// IsUpToDate checks if an artifact's input hash matches the lockfile.
func (lf *LockFile) IsUpToDate(artifactID, inputHash string) bool {
	entry, ok := lf.Artifacts[artifactID]
	if !ok {
		return false
	}
	return entry.InputHash == inputHash
}

// CacheDir returns the .sc-cache directory path.
func CacheDir(projectDir string) string {
	return filepath.Join(projectDir, ".sc-cache")
}

// ReadCached reads a cached artifact output.
func ReadCached(projectDir, artifactID string) (string, error) {
	path := filepath.Join(CacheDir(projectDir), artifactID)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteCached writes an artifact output to the cache.
func WriteCached(projectDir, artifactID, content string) error {
	dir := CacheDir(projectDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, artifactID), []byte(content), 0o644)
}
