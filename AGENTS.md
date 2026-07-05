# AGENTS.md — YAML Forge

## Overview
YAML Forge is a comprehensive YAML processing toolkit in Go. It provides parsing, querying, diffing, merging, converting, linting, formatting, filtering, validation, and statistics for YAML documents.

## Project Structure
- `cmd/yamlforge/` — CLI entry point using Cobra
- `internal/models/` — Core data types (Node, PathSegment, LintIssue, YAMLStats, etc.)
- `internal/parser/` — YAML parsing with comment preservation
- `internal/query/` — YAMLPath query engine
- `internal/diff/` — Semantic YAML diff engine
- `internal/merge/` — Deep/shallow merge strategies
- `internal/lint/` — Linting rules (raw text + tree analysis)
- `internal/format/` — Pretty-print, sort keys, minify
- `internal/convert/` — YAML ↔ JSON ↔ TOML conversion
- `internal/filter/` — Multi-document filtering
- `internal/validate/` — Schema validation engine
- `internal/stats/` — Document statistics
- `tests/` — Test packages (one per internal package)

## Key Design Decisions

### yaml.v3 DocumentNode
When using `yaml.Unmarshal`, the root node is always a `yaml.DocumentNode` wrapping the actual content. All functions that traverse `yaml.Node` trees must handle this:

```go
if node.Kind == yaml.DocumentNode {
    for _, child := range node.Content {
        // process child (the actual mapping/sequence/scalar)
    }
    return
}
```

### Lint Rules
Lint rules operate at two levels:
1. **Raw text** — for checks like duplicate keys and truthy values (since yaml.v3 silently deduplicates keys and converts `yes`/`no` to `true`/`false`)
2. **Parsed tree** — for structural checks

### Query Engine
The query engine uses dot-notation paths:
- `server.port` — nested keys
- `items[0]` — array indices
- `items[*]` — wildcards

### Encoding
The `yaml.Encoder` requires a `DocumentNode` as input. When encoding back from a parsed node, wrap it:

```go
if node.Kind != yaml.DocumentNode {
    node = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}}
}
```

## Testing
Tests are in `tests/<pkg>/<pkg>_test.go` using the same package name (not `_test` suffix). This gives access to unexported functions.

Run all tests:
```bash
go test ./...
```

## Dependencies
- `gopkg.in/yaml.v3` — YAML parsing/encoding
- `github.com/spf13/cobra` — CLI framework
