# YAML Forge

A comprehensive YAML processing toolkit in Go — parse, query, diff, merge, convert, lint, format, filter, validate, and analyze YAML documents.

## Features

| Feature | Description |
|---------|-------------|
| **Parse** | Parse YAML with comment preservation, anchors, aliases, multi-document support |
| **Query** | YAMLPath query engine — dot-notation, array indices, wildcards |
| **Diff** | Semantic YAML diff with field-level change tracking |
| **Merge** | Deep/shallow/replace/append/keep merge strategies |
| **Convert** | YAML ↔ JSON ↔ TOML conversion |
| **Lint** | 6 built-in rules: duplicate keys, truthy values, trailing spaces, consistent indentation, line length, empty keys |
| **Format** | Pretty-print, sort keys, minify, custom indentation |
| **Filter** | Multi-document filtering by content, index, or regex |
| **Validate** | Schema validation — required fields, type checking, enums, min/max length |
| **Statistics** | Node counting, depth analysis, key frequency, line count |
| **CLI** | 14+ commands for all operations |

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/yamlforge/cmd/yamlforge@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/yamlforge.git
cd yamlforge
go build -o yamlforge ./cmd/yamlforge
```

## Quick Start

### CLI Usage

```bash
# Pretty-print a YAML file
yamlforge fmt config.yaml

# Lint for common issues
yamlforge lint config.yaml

# Query a specific value
yamlforge query config.yaml "server.port"

# Diff two YAML files
yamlforge diff old.yaml new.yaml

# Convert YAML to JSON
yamlforge convert config.yaml --to json

# Validate against a schema
yamlforge validate config.yaml --schema schema.yaml

# Show document statistics
yamlforge stats config.yaml

# Sort keys alphabetically
yamlforge sort config.yaml

# Merge two files (deep merge)
yamlforge merge base.yaml override.yaml

# Flatten nested YAML
yamlforge flatten config.yaml

# Count documents in a multi-doc file
yamlforge count config.yaml

# List all keys
yamlforge keys config.yaml

# Infer schema from existing YAML
yamlforge schema config.yaml

# Print raw YAML content
yamlforge print config.yaml
```

### Library Usage

```go
package main

import (
    "fmt"
    "github.com/EdgarOrtegaRamirez/yamlforge/internal/parser"
    "github.com/EdgarOrtegaRamirez/yamlforge/internal/query"
    "github.com/EdgarOrtegaRamirez/yamlforge/internal/lint"
    "github.com/EdgarOrtegaRamirez/yamlforge/internal/format"
)

func main() {
    // Parse YAML
    p := parser.New()
    docs, _ := p.ParseFile("config.yaml")
    
    // Query
    engine := query.NewEngine()
    value, _ := engine.QueryValue(docs[0], "server.port")
    fmt.Println("Port:", value)
    
    // Lint
    linter := lint.NewEngine()
    report, _ := linter.LintFile("config.yaml")
    for _, issue := range report.Issues {
        fmt.Printf("%s: %s (line %d)\n", issue.Rule, issue.Message, issue.Line)
    }
    
    // Format
    formatted, _ := format.SortKeys(data)
    fmt.Println(string(formatted))
}
```

## Architecture

```
yamlforge/
├── cmd/yamlforge/       # CLI entry point (Cobra)
├── internal/
│   ├── models/          # Core data types (Node, PathSegment, LintIssue, YAMLStats)
│   ├── parser/          # YAML parsing with comment preservation
│   ├── query/           # YAMLPath query engine
│   ├── diff/            # Semantic YAML diff
│   ├── merge/           # Deep/shallow merge strategies
│   ├── lint/            # Linting rules (raw text + tree analysis)
│   ├── format/          # Pretty-print, sort keys, minify
│   ├── convert/         # YAML ↔ JSON ↔ TOML conversion
│   ├── filter/          # Multi-document filtering
│   ├── validate/        # Schema validation engine
│   └── stats/           # Document statistics
└── tests/               # Test packages (11 test suites)
```

## Lint Rules

| Rule | Severity | Description |
|------|----------|-------------|
| `no-duplicate-keys` | error | Detects duplicate keys in the same scope |
| `truthy-values` | warning | Flags ambiguous values (yes/no/on/off) |
| `no-trailing-spaces` | warning | Detects trailing whitespace |
| `consistent-indentation` | warning | Checks indentation consistency |
| `line-length` | warning | Flags lines exceeding 200 characters |
| `no-empty-keys` | warning | Detects empty mapping keys |

## Merge Strategies

| Strategy | Description |
|----------|-------------|
| `deep` | Recursively merge nested objects (default) |
| `replace` | Replace entire value from second file |
| `append` | Append sequences together |
| `keep` | Keep original value, skip conflicts |

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./tests/lint/ -v
go test ./tests/query/ -v
go test ./tests/cli/ -v
```

## License

MIT
