package ir

// IntermediateRepr is the normalized representation all spec plugins parse into.
type IntermediateRepr struct {
	Operations []Operation       `json:"operations,omitempty"`
	Types      []TypeDef         `json:"types,omitempty"`
	Auth       []AuthScheme      `json:"auth,omitempty"`
	Groups     []Group           `json:"groups,omitempty"`
	Structure  *ProjectStructure `json:"structure,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Operation represents an endpoint, command, or RPC.
type Operation struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Method      string      `json:"method,omitempty"` // HTTP method or empty for CLI
	Path        string      `json:"path,omitempty"`   // HTTP path or command path
	Parameters  []Parameter `json:"parameters,omitempty"`
	RequestBody *TypeRef    `json:"requestBody,omitempty"`
	Responses   []Response  `json:"responses,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Deprecated  bool        `json:"deprecated,omitempty"`
	Auth        []string    `json:"auth,omitempty"` // references to AuthScheme IDs
	// CLI-specific
	Aliases     []string `json:"aliases,omitempty"`
	RawHelpText string   `json:"rawHelpText,omitempty"`
}

// Parameter represents a flag, query param, path param, or header.
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in,omitempty"` // query, path, header, cookie, flag, argument
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Type        string `json:"type,omitempty"`
	Default     string `json:"default,omitempty"`
	Shorthand   string `json:"shorthand,omitempty"` // CLI short flag
}

// TypeDef represents a schema, message type, or complex value type.
type TypeDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Fields      []TypeField `json:"fields,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

// TypeField is a field within a TypeDef.
type TypeField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// TypeRef references a type by name, used for request/response bodies.
type TypeRef struct {
	TypeName    string `json:"typeName,omitempty"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

// Response represents an HTTP response or command output.
type Response struct {
	StatusCode  string  `json:"statusCode"`
	Description string  `json:"description,omitempty"`
	Body        *TypeRef `json:"body,omitempty"`
}

// AuthScheme represents an authentication method.
type AuthScheme struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // apiKey, http, oauth2, openIdConnect
	Name        string `json:"name,omitempty"`
	In          string `json:"in,omitempty"` // header, query, cookie
	Scheme      string `json:"scheme,omitempty"` // bearer, basic
	Description string `json:"description,omitempty"`
}

// Group organizes operations by resource, tag, or subcommand tree.
type Group struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Operations  []string `json:"operations,omitempty"` // operation IDs
}

// ProjectStructure holds codebase scan results (codebase plugin only).
type ProjectStructure struct {
	FileTree    []FileEntry  `json:"fileTree,omitempty"`
	Stack       *StackInfo   `json:"stack,omitempty"`
	EntryPoints []string     `json:"entryPoints,omitempty"`
	ConfigFiles []ConfigFile `json:"configFiles,omitempty"`
	Docs        []DocFile    `json:"docs,omitempty"`
	KeyFiles    []KeyFile    `json:"keyFiles,omitempty"`
}

// FileEntry is a file in the project tree.
type FileEntry struct {
	Path  string `json:"path"`
	IsDir bool   `json:"isDir,omitempty"`
	Size  int64  `json:"size,omitempty"`
}

// StackInfo describes the project's technology stack.
type StackInfo struct {
	Languages    []string          `json:"languages,omitempty"`
	Frameworks   []string          `json:"frameworks,omitempty"`
	BuildTools   []string          `json:"buildTools,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Scripts      map[string]string `json:"scripts,omitempty"`
}

// ConfigFile is a parsed config file from the project.
type ConfigFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// DocFile is an existing documentation file.
type DocFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// KeyFile is an important source file identified by the scanner.
type KeyFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Role    string `json:"role,omitempty"` // entrypoint, routes, schema, test-setup
}

// Merge combines another IR into this one.
func (ir *IntermediateRepr) Merge(other *IntermediateRepr) {
	if other == nil {
		return
	}
	ir.Operations = append(ir.Operations, other.Operations...)
	ir.Types = append(ir.Types, other.Types...)
	ir.Auth = append(ir.Auth, other.Auth...)
	ir.Groups = append(ir.Groups, other.Groups...)
	if other.Structure != nil {
		if ir.Structure == nil {
			ir.Structure = other.Structure
		} else {
			ir.Structure.FileTree = append(ir.Structure.FileTree, other.Structure.FileTree...)
			ir.Structure.EntryPoints = append(ir.Structure.EntryPoints, other.Structure.EntryPoints...)
			ir.Structure.ConfigFiles = append(ir.Structure.ConfigFiles, other.Structure.ConfigFiles...)
			ir.Structure.Docs = append(ir.Structure.Docs, other.Structure.Docs...)
			ir.Structure.KeyFiles = append(ir.Structure.KeyFiles, other.Structure.KeyFiles...)
		}
	}
	if ir.Metadata == nil {
		ir.Metadata = make(map[string]string)
	}
	for k, v := range other.Metadata {
		ir.Metadata[k] = v
	}
}
