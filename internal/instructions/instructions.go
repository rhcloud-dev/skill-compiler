package instructions

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Instructions represents a parsed COMPILER_INSTRUCTIONS.md file.
type Instructions struct {
	Frontmatter Frontmatter
	Sections    map[string]string // H1 heading -> content
	RawBody     string
}

// Frontmatter holds all YAML frontmatter fields.
type Frontmatter struct {
	Name      string          `yaml:"name"`
	Spec      yaml.Node       `yaml:"spec"`      // string, object, or array
	Out       string          `yaml:"out"`        // default: ./sc-out/
	Artifacts map[string]Artifact `yaml:"artifacts"` // per-artifact toggles
	Skill     SkillConfig     `yaml:"skill"`
	Provider  ProviderConfig  `yaml:"provider"`
}

// SpecSource represents a resolved spec source.
type SpecSource struct {
	// For file paths
	Path string `yaml:"path,omitempty"`
	// For URLs
	URL string `yaml:"url,omitempty"`
	// For shell commands
	Command string `yaml:"command,omitempty"`
	// Type: openapi, cli, codebase
	Type string `yaml:"type,omitempty"`
	// CLI-specific
	Binary   string   `yaml:"binary,omitempty"`
	HelpFlag string   `yaml:"help-flag,omitempty"`
	MaxDepth int      `yaml:"max-depth,omitempty"`
	Exclude  []string `yaml:"exclude,omitempty"`
	// Codebase-specific
	MaxFiles int      `yaml:"max-files,omitempty"`
	Include  []string `yaml:"include,omitempty"`
}

// Artifact controls per-artifact settings.
type Artifact struct {
	Enabled  *bool  `yaml:"enabled,omitempty"`
	Filename string `yaml:"filename,omitempty"`
}

// IsEnabled returns whether this artifact is enabled (default true).
func (a Artifact) IsEnabled() bool {
	if a.Enabled == nil {
		return true
	}
	return *a.Enabled
}

// SkillConfig holds skill metadata for the generated SKILL.md.
type SkillConfig struct {
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	Env           []string          `yaml:"env,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`
}

// ProviderConfig holds per-project LLM provider overrides.
type ProviderConfig struct {
	Provider string `yaml:"provider,omitempty"`
	Model    string `yaml:"model,omitempty"`
	APIKey   string `yaml:"api-key,omitempty"`
	BaseURL  string `yaml:"base-url,omitempty"`
}

// Parse reads and parses a COMPILER_INSTRUCTIONS.md file.
func Parse(path string) (*Instructions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading instructions file: %w", err)
	}
	return ParseBytes(data)
}

// ParseBytes parses instructions from raw bytes.
func ParseBytes(data []byte) (*Instructions, error) {
	content := string(data)

	fm, body, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var frontmatter Frontmatter
	if err := yaml.Unmarshal([]byte(fm), &frontmatter); err != nil {
		return nil, fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	if frontmatter.Name == "" {
		return nil, fmt.Errorf("frontmatter missing required field: name")
	}

	if frontmatter.Out == "" {
		frontmatter.Out = "./sc-out/"
	}

	sections := extractSections(body)

	return &Instructions{
		Frontmatter: frontmatter,
		Sections:    sections,
		RawBody:     body,
	}, nil
}

// ResolveSpecSources converts the raw YAML spec node into typed SpecSource(s).
func (inst *Instructions) ResolveSpecSources() ([]SpecSource, error) {
	node := &inst.Frontmatter.Spec
	if node.IsZero() {
		// Default to openapi.yaml
		return []SpecSource{{Path: "./openapi.yaml"}}, nil
	}
	return resolveSpecNode(node)
}

func resolveSpecNode(node *yaml.Node) ([]SpecSource, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		// String form: file path
		return []SpecSource{{Path: node.Value}}, nil

	case yaml.MappingNode:
		// Single object
		var src SpecSource
		if err := node.Decode(&src); err != nil {
			return nil, fmt.Errorf("parsing spec source object: %w", err)
		}
		return []SpecSource{src}, nil

	case yaml.SequenceNode:
		// Array of sources
		var sources []SpecSource
		for _, child := range node.Content {
			resolved, err := resolveSpecNode(child)
			if err != nil {
				return nil, err
			}
			sources = append(sources, resolved...)
		}
		return sources, nil

	default:
		return nil, fmt.Errorf("unsupported spec format (YAML kind: %d)", node.Kind)
	}
}

// extractFrontmatter splits on --- delimiters and returns frontmatter YAML and body.
func extractFrontmatter(content string) (string, string, error) {
	// Must start with ---
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return "", "", fmt.Errorf("instructions file must start with YAML frontmatter (---)")
	}

	// Find second ---
	rest := trimmed[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", "", fmt.Errorf("instructions file missing closing frontmatter delimiter (---)")
	}

	fm := strings.TrimSpace(rest[:idx])
	body := strings.TrimSpace(rest[idx+4:])
	return fm, body, nil
}

// extractSections splits the markdown body on H1 headings into named sections.
func extractSections(body string) map[string]string {
	sections := make(map[string]string)
	if body == "" {
		return sections
	}

	lines := strings.Split(body, "\n")
	var currentSection string
	var currentContent []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			// Save previous section
			if currentSection != "" {
				sections[currentSection] = strings.TrimSpace(strings.Join(currentContent, "\n"))
			}
			currentSection = strings.TrimSpace(line[2:])
			currentContent = nil
		} else {
			currentContent = append(currentContent, line)
		}
	}

	// Save last section
	if currentSection != "" {
		sections[currentSection] = strings.TrimSpace(strings.Join(currentContent, "\n"))
	}

	return sections
}

// Validate checks the instructions for common issues, returning warnings.
func (inst *Instructions) Validate() []string {
	var warnings []string
	if _, ok := inst.Sections["Product"]; !ok {
		warnings = append(warnings, "missing recommended section: # Product")
	}
	return warnings
}

// EnvPrefix derives the env var prefix from the name field.
// e.g., "my-app" -> "MY_APP"
func (inst *Instructions) EnvPrefix() string {
	return strings.ToUpper(strings.ReplaceAll(inst.Frontmatter.Name, "-", "_"))
}
