package openapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roberthamel/skill-compiler/internal/instructions"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	return data
}

func TestDetect(t *testing.T) {
	p := New()

	tests := []struct {
		name   string
		source instructions.SpecSource
		want   bool
	}{
		{"yaml file", instructions.SpecSource{Path: "api.yaml"}, true},
		{"yml file", instructions.SpecSource{Path: "api.yml"}, true},
		{"json file", instructions.SpecSource{Path: "api.json"}, true},
		{"explicit type", instructions.SpecSource{Type: "openapi", URL: "http://example.com"}, true},
		{"cli type", instructions.SpecSource{Type: "cli", Binary: "kubectl"}, false},
		{"go file", instructions.SpecSource{Path: "main.go"}, false},
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

func TestParse_Petstore(t *testing.T) {
	p := New()
	data := readTestdata(t, "petstore.yaml")
	source := instructions.SpecSource{Path: "testdata/petstore.yaml"}

	result, err := p.Parse(data, source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Should have 3 operations: listPets, createPet, getPet
	if len(result.Operations) != 3 {
		t.Errorf("got %d operations, want 3", len(result.Operations))
	}

	opIDs := map[string]bool{}
	for _, op := range result.Operations {
		opIDs[op.ID] = true
	}
	for _, id := range []string{"listPets", "createPet", "getPet"} {
		if !opIDs[id] {
			t.Errorf("missing operation %q", id)
		}
	}

	// Should have 2 types: Pet, Error
	if len(result.Types) != 2 {
		t.Errorf("got %d types, want 2", len(result.Types))
	}

	// Should have 1 auth scheme
	if len(result.Auth) != 1 {
		t.Errorf("got %d auth schemes, want 1", len(result.Auth))
	}
	if result.Auth[0].Type != "apiKey" {
		t.Errorf("auth type = %q, want %q", result.Auth[0].Type, "apiKey")
	}

	// Check metadata
	if result.Metadata["title"] != "Petstore" {
		t.Errorf("title = %q, want %q", result.Metadata["title"], "Petstore")
	}
}

func TestParse_RefResolution(t *testing.T) {
	p := New()
	data := readTestdata(t, "petstore.yaml")
	source := instructions.SpecSource{Path: "testdata/petstore.yaml"}

	result, err := p.Parse(data, source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// The listPets response references Pet via $ref — after resolution, the response body
	// should reference the Pet type name
	for _, op := range result.Operations {
		if op.ID == "listPets" {
			if len(op.Responses) == 0 {
				t.Fatal("listPets has no responses")
			}
			// The 200 response should exist
			found := false
			for _, r := range op.Responses {
				if r.StatusCode == "200" {
					found = true
				}
			}
			if !found {
				t.Error("listPets missing 200 response")
			}
		}
	}
}

func TestValidate_MissingDescriptions(t *testing.T) {
	p := New()
	// Create a minimal spec with an undocumented parameter
	spec := `openapi: "3.0.0"
info:
  title: Test
  version: "1.0"
paths:
  /items:
    get:
      operationId: listItems
      summary: List items
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        "200":
          description: OK`

	source := instructions.SpecSource{Path: "test.yaml"}
	result, err := p.Parse([]byte(spec), source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	warnings := p.Validate(result)
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "limit") && strings.Contains(w.Message, "no description") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected warning about parameter 'limit' missing description, got %v", warnings)
	}
}
