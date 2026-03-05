# Examples

## COMPILER_INSTRUCTIONS.md

A fully annotated example showing all available frontmatter fields and recommended markdown sections.

Use it as a starting point:

```sh
cp examples/COMPILER_INSTRUCTIONS.md ./COMPILER_INSTRUCTIONS.md
# Edit to match your API/CLI, then:
sc generate
```

Or scaffold from a spec and compare with this example:

```sh
sc init --name my-api --spec ./openapi.yaml
# Compare with examples/COMPILER_INSTRUCTIONS.md to see what sections to add
```

## Sections reference

| Section | Purpose |
|---------|---------|
| `# Product` | What the tool does, who it's for, key capabilities |
| `# Workflows` | Multi-step sequences agents will commonly perform |
| `# Guardrails` | Safety rules, rate limits, irreversible actions to watch for |
| `# Conventions` | Naming patterns, date formats, pagination, ID formats |
| `# Examples` | Concrete request/response samples |
| `# Common patterns` | Reusable tips (caching, batching, idempotency) |
