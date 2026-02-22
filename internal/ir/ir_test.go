package ir

import (
	"testing"

	"github.com/roberthamel/skill-compiler/internal/instructions"
)

func TestMerge(t *testing.T) {
	a := &IntermediateRepr{
		Operations: []Operation{{ID: "op1", Name: "Op 1"}},
		Types:      []TypeDef{{Name: "TypeA"}},
		Auth:       []AuthScheme{{ID: "auth1", Type: "apiKey"}},
		Metadata:   map[string]string{"key1": "val1"},
	}
	b := &IntermediateRepr{
		Operations: []Operation{{ID: "op2", Name: "Op 2"}},
		Types:      []TypeDef{{Name: "TypeB"}},
		Auth:       []AuthScheme{{ID: "auth2", Type: "http"}},
		Metadata:   map[string]string{"key2": "val2"},
	}

	a.Merge(b)

	if len(a.Operations) != 2 {
		t.Errorf("got %d operations, want 2", len(a.Operations))
	}
	if len(a.Types) != 2 {
		t.Errorf("got %d types, want 2", len(a.Types))
	}
	if len(a.Auth) != 2 {
		t.Errorf("got %d auth schemes, want 2", len(a.Auth))
	}
	if a.Metadata["key1"] != "val1" || a.Metadata["key2"] != "val2" {
		t.Errorf("metadata = %v, want both keys", a.Metadata)
	}
}

func TestMerge_Nil(t *testing.T) {
	a := &IntermediateRepr{Operations: []Operation{{ID: "op1"}}}
	a.Merge(nil)
	if len(a.Operations) != 1 {
		t.Error("merge with nil should not change IR")
	}
}

// mockPlugin is a test plugin that always returns a fixed IR.
type mockPlugin struct {
	name      string
	detectFn  func(instructions.SpecSource) bool
	ir        *IntermediateRepr
	warnings  []Warning
	fetchData []byte
}

func (m *mockPlugin) Name() string                          { return m.name }
func (m *mockPlugin) Detect(s instructions.SpecSource) bool { return m.detectFn(s) }
func (m *mockPlugin) Fetch(_ instructions.SpecSource) ([]byte, error) {
	return m.fetchData, nil
}
func (m *mockPlugin) Parse(_ []byte, _ instructions.SpecSource) (*IntermediateRepr, error) {
	return m.ir, nil
}
func (m *mockPlugin) Validate(_ *IntermediateRepr) []Warning { return m.warnings }

func TestRegistry_ProcessSources(t *testing.T) {
	plugin := &mockPlugin{
		name:      "mock",
		detectFn:  func(s instructions.SpecSource) bool { return s.Type == "mock" },
		ir:        &IntermediateRepr{Operations: []Operation{{ID: "mock_op"}}},
		fetchData: []byte("data"),
	}

	reg := NewRegistry()
	reg.Register(plugin)

	sources := []instructions.SpecSource{{Type: "mock"}}
	result, warnings, err := reg.ProcessSources(sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("got %d warnings, want 0", len(warnings))
	}
	if len(result.Operations) != 1 || result.Operations[0].ID != "mock_op" {
		t.Errorf("operations = %+v, want [mock_op]", result.Operations)
	}
}

func TestRegistry_Detect(t *testing.T) {
	openapi := &mockPlugin{
		name:     "openapi",
		detectFn: func(s instructions.SpecSource) bool { return s.Type == "openapi" },
	}
	cli := &mockPlugin{
		name:     "cli",
		detectFn: func(s instructions.SpecSource) bool { return s.Type == "cli" },
	}

	reg := NewRegistry()
	reg.Register(openapi)
	reg.Register(cli)

	p, err := reg.Detect(instructions.SpecSource{Type: "cli"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "cli" {
		t.Errorf("detected plugin = %q, want %q", p.Name(), "cli")
	}

	_, err = reg.Detect(instructions.SpecSource{Type: "unknown"})
	if err == nil {
		t.Error("expected error for unknown source type")
	}
}
