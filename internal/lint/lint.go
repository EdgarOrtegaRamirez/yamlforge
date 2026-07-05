// Package lint provides YAML linting with configurable rules.
package lint

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"gopkg.in/yaml.v3"
)

// Rule represents a lint rule.
type Rule interface {
	Name() string
	Description() string
	CheckFile(path string, data []byte) []models.LintIssue
}

// Engine runs lint rules against YAML files.
type Engine struct {
	rules []Rule
}

// NewEngine creates a new lint engine with default rules.
func NewEngine() *Engine {
	return &Engine{
		rules: []Rule{
			&NoTrailingSpaces{},
			&ConsistentIndentation{},
			&NoDuplicateKeys{},
			&TruthyValues{},
			&LineLength{MaxLen: 200},
			&NoEmptyMappingKeys{},
		},
	}
}

// AddRule adds a custom rule to the engine.
func (e *Engine) AddRule(rule Rule) {
	e.rules = append(e.rules, rule)
}

// LintFile lints a YAML file and returns a report.
func (e *Engine) LintFile(path string) (*models.LintReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	report := &models.LintReport{File: path}

	// Check for YAML parse errors
	report.Issues = append(report.Issues, CheckParseError(path, data)...)

	// Run all rules
	for _, rule := range e.rules {
		issues := rule.CheckFile(path, data)
		report.Issues = append(report.Issues, issues...)
	}

	return report, nil
}

// CheckParseError checks for YAML syntax errors in raw data.
func CheckParseError(path string, data []byte) []models.LintIssue {
	var issues []models.LintIssue
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		issues = append(issues, models.LintIssue{
			Line:     0,
			Severity: "error",
			Rule:     "syntax",
			Message:  err.Error(),
		})
	}
	return issues
}

// LintBytes lints YAML bytes.
func (e *Engine) LintBytes(data []byte, filename string) (*models.LintReport, error) {
	report := &models.LintReport{File: filename}

	for _, rule := range e.rules {
		issues := rule.CheckFile(filename, data)
		report.Issues = append(report.Issues, issues...)
	}

	return report, nil
}

// NoTrailingSpaces checks for trailing whitespace.
type NoTrailingSpaces struct{}

func (r *NoTrailingSpaces) Name() string        { return "no-trailing-spaces" }
func (r *NoTrailingSpaces) Description() string { return "Checks for trailing whitespace" }

func (r *NoTrailingSpaces) CheckFile(path string, data []byte) []models.LintIssue {
	var issues []models.LintIssue
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.TrimRight(line, " \t") != line && strings.TrimSpace(line) != "" {
			issues = append(issues, models.LintIssue{
				Line:     i + 1,
				Severity: "warning",
				Rule:     r.Name(),
				Message:  "trailing whitespace",
			})
		}
	}
	return issues
}

// ConsistentIndentation checks for consistent indentation.
type ConsistentIndentation struct{}

func (r *ConsistentIndentation) Name() string        { return "consistent-indentation" }
func (r *ConsistentIndentation) Description() string { return "Checks for consistent indentation" }

func (r *ConsistentIndentation) CheckFile(path string, data []byte) []models.LintIssue {
	var issues []models.LintIssue
	lines := strings.Split(string(data), "\n")
	indentSize := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if indent > 0 && indentSize == 0 {
			indentSize = indent
		}
		if indentSize > 0 && indent%indentSize != 0 {
			issues = append(issues, models.LintIssue{
				Severity: "warning",
				Rule:     r.Name(),
				Message:  fmt.Sprintf("inconsistent indentation (expected multiple of %d)", indentSize),
			})
			break
		}
	}
	return issues
}

// NoDuplicateKeys checks for duplicate keys using raw line parsing.
type NoDuplicateKeys struct{}

func (r *NoDuplicateKeys) Name() string        { return "no-duplicate-keys" }
func (r *NoDuplicateKeys) Description() string { return "Checks for duplicate mapping keys" }

func (r *NoDuplicateKeys) CheckFile(path string, data []byte) []models.LintIssue {
	var issues []models.LintIssue
	lines := strings.Split(string(data), "\n")

	type keyInfo struct {
		line  int
		depth int
	}
	seen := make(map[string]keyInfo)
	currentDepth := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || trimmed == "---" {
			continue
		}

		// Calculate indentation depth
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Check if this is a key (contains : and is not in a value)
		if idx := strings.Index(trimmed, ":"); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])

			// Skip if it looks like a URL or value
			if strings.HasPrefix(key, "http") || strings.Contains(key, "://") {
				continue
			}

			// Skip array items
			if strings.HasPrefix(trimmed, "- ") {
				continue
			}

			// Create a scope key based on depth
			scopeKey := fmt.Sprintf("%d:%s", indent, key)

			if prev, ok := seen[scopeKey]; ok {
				issues = append(issues, models.LintIssue{
					Line:     i + 1,
					Severity: "error",
					Rule:     r.Name(),
					Message:  fmt.Sprintf("duplicate key %q (first at line %d)", key, prev.line),
				})
			} else {
				seen[scopeKey] = keyInfo{line: i + 1, depth: indent}
			}
		}
		_ = currentDepth
	}
	return issues
}

// TruthyValues checks for ambiguous truthy values using raw text.
type TruthyValues struct{}

func (r *TruthyValues) Name() string        { return "truthy-values" }
func (r *TruthyValues) Description() string { return "Checks for ambiguous truthy values (yes/no/on/off)" }

var truthyPattern = regexp.MustCompile(`:\s*(yes|no|on|off|Yes|No|YES|NO|ON|OFF)\s*$`)

func (r *TruthyValues) CheckFile(path string, data []byte) []models.LintIssue {
	var issues []models.LintIssue
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || trimmed == "---" {
			continue
		}

		// Skip array items
		if strings.HasPrefix(trimmed, "- ") {
			continue
		}

		if matches := truthyPattern.FindStringSubmatch(line); len(matches) > 1 {
			issues = append(issues, models.LintIssue{
				Line:     i + 1,
				Severity: "warning",
				Rule:     r.Name(),
				Message:  fmt.Sprintf("ambiguous value %q (use explicit true/false)", matches[1]),
			})
		}
	}
	return issues
}

// LineLength checks for excessively long lines.
type LineLength struct {
	MaxLen int
}

func (r *LineLength) Name() string        { return "line-length" }
func (r *LineLength) Description() string { return "Checks for excessively long lines" }

func (r *LineLength) CheckFile(path string, data []byte) []models.LintIssue {
	var issues []models.LintIssue
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if len(line) > r.MaxLen {
			issues = append(issues, models.LintIssue{
				Line:     i + 1,
				Severity: "warning",
				Rule:     r.Name(),
				Message:  fmt.Sprintf("line too long (%d > %d)", len(line), r.MaxLen),
			})
		}
	}
	return issues
}

// NoEmptyMappingKeys checks for empty mapping keys.
type NoEmptyMappingKeys struct{}

func (r *NoEmptyMappingKeys) Name() string        { return "no-empty-keys" }
func (r *NoEmptyMappingKeys) Description() string { return "Checks for empty mapping keys" }

func (r *NoEmptyMappingKeys) CheckFile(path string, data []byte) []models.LintIssue {
	var issues []models.LintIssue
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || trimmed == "---" {
			continue
		}

		// Check for empty key (just a colon)
		if trimmed == ":" || strings.HasPrefix(trimmed, ": ") {
			issues = append(issues, models.LintIssue{
				Line:     i + 1,
				Severity: "warning",
				Rule:     r.Name(),
				Message:  "empty mapping key",
			})
		}

		// Check for key: "" (empty string value)
		if idx := strings.Index(trimmed, ":"); idx > 0 {
			key := trimmed[:idx]
			val := strings.TrimSpace(trimmed[idx+1:])
			// This is valid, just note it
			_ = key
			_ = val
		}
	}
	return issues
}

// ParseLineCount counts logical lines.
func ParseLineCount(data []byte) int {
	return len(strings.Split(string(data), "\n"))
}

// GetIndentLevel returns the indent level of a line.
func GetIndentLevel(line string) int {
	indent := 0
	for _, c := range line {
		if c == ' ' {
			indent++
		} else if c == '\t' {
			indent += 2 // treat tab as 2 spaces
		} else {
			break
		}
	}
	return indent
}

// ParseInt safely parses an integer string.
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
