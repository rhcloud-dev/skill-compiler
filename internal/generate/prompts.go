package generate

// System prompt templates for each artifact type.

const SkillPrompt = `You are generating a SKILL.md file for an Agent Skills spec-compliant skill directory.

Your output must be a complete SKILL.md file with:
1. YAML frontmatter (between --- delimiters) containing:
   - name: (provided, must match exactly)
   - description: (max 1024 chars, describe what the skill does and when to use it)
   - Any additional metadata fields provided (license, compatibility, metadata, allowed-tools)

2. Markdown body (UNDER 500 lines) structured for progressive disclosure:
   - ## Configuration — environment variables, authentication setup
   - ## Core Concepts — mental model for the tool
   - ## Key Operations — most important operations with brief usage
   - ## Value Formats — important data types and formats
   - ## Best Practices — guardrails, conventions, common pitfalls
   - ## File References — pointers to references/ and scripts/ for details

The body should be optimized for an AI agent to quickly understand and use the tool.
Keep it concise but comprehensive. Use relative file references (e.g., references/reference.md).
Do NOT include raw API specs — that goes in references/.
Do NOT exceed 500 lines in the body.`

const ReferencePrompt = `You are generating a reference.md file — an exhaustive command/endpoint reference.

Your output must be a complete markdown document listing EVERY operation with:
- Full path/command syntax
- All parameters, flags, arguments with types and descriptions
- Request/response body shapes (for APIs)
- Error codes and their meanings
- Authentication requirements

Organize by resource/domain area. Use consistent formatting.
Be thorough — this is the complete reference an agent loads on demand.`

const ExamplesPrompt = `You are generating an examples.md file — worked multi-step workflow examples.

Your output must show realistic, end-to-end workflows that combine multiple operations.
Each example should:
- Have a clear title describing the goal
- Show the complete sequence of operations
- Include realistic sample data
- Explain what each step does and why
- Show expected responses/outputs

Focus on the most common workflows agents would perform.
Pull from any provided workflow descriptions, common patterns, and domain knowledge.`

const ScriptsPrompt = `You are generating executable shell scripts for a skill's scripts/ directory.

Generate as many useful scripts as you can identify from the spec and instructions.
Each script should:
- Start with #!/bin/bash (or #!/bin/sh)
- Have a comment header explaining: purpose, required env vars, usage
- Be directly executable by an agent
- Combine multiple operations into single scripts where useful

Types of scripts to generate:
- health-check.sh: Validate connectivity and auth
- discover.sh: List available resources
- Custom workflow scripts that save multi-step operations

Output format: Output each script as a code block with the filename as the info string.
Example:
` + "```health-check.sh" + `
#!/bin/bash
# Purpose: Validate API connectivity and authentication
# Env vars: MY_APP_API_URL, MY_APP_API_KEY
# Usage: ./health-check.sh

curl -s -o /dev/null -w "%{http_code}" ...
` + "```" + `

Generate ALL useful scripts you can identify.`

const LlmsTxtPrompt = `You are generating an llms.txt file — a brief product overview (~500 tokens).

Your output must be a concise description including:
- What the tool/service does (1-2 sentences)
- Key capabilities as bullet points
- Links to other documentation files

This is the lightest-weight description for quick context.
Target approximately 500 tokens total.`

const LlmsAPITxtPrompt = `You are generating an llms-api.txt file — a concise interface reference (~2-4K tokens).

Your output must include:
- Quick start (authentication, base URL)
- Every operation as a ONE-LINE summary (method + path + brief description)
- Common patterns (pagination, filtering, error handling)
- Error codes table

Be concise but complete — every operation should appear.
Target approximately 2000-4000 tokens.`

const LlmsFullTxtPrompt = `You are generating an llms-full.txt file — complete documentation (~5-15K tokens).

Your output must include:
- Overview and core concepts
- Authentication and configuration
- All operations with full details (parameters, request/response shapes, examples)
- Worked examples of common workflows
- Error handling and troubleshooting
- Best practices and conventions

This is the most detailed single-file documentation.
Target approximately 5000-15000 tokens.`

const ChangelogPrompt = `You are generating a CHANGELOG.md entry by comparing previous and current specs/instructions.

Generate a dated changelog entry with these sections (omit empty sections):
### Added — New operations, features, or capabilities
### Changed — Modified parameters, updated behavior, changed defaults
### Deprecated — Operations or features marked for removal
### Removed — Operations or features that no longer exist
### Instructions — Changes to guidance, workflows, or guardrails

Be specific: list operation names, parameter changes, before/after values.
If this is the first generation (no previous artifacts), create an "Initial generation" entry.`

const InitPrompt = `You are generating a COMPILER_INSTRUCTIONS.md file from a spec.

Your output must be a complete COMPILER_INSTRUCTIONS.md with:
1. YAML frontmatter (between --- delimiters) with the provided name and spec configuration
2. Markdown body with draft sections that the user should review and customize:
   - # Product — What the tool does, target users, key value props
   - # Workflows — Common multi-step workflows agents will perform
   - # Guardrails — Safety rules, rate limits, things to avoid
   - # Conventions — Naming patterns, value formats, common patterns

Base the draft content on what you can infer from the spec.
Mark sections that need human review with <!-- REVIEW: ... --> comments.`
