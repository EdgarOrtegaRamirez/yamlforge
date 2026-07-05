// Package validate provides YAML schema validation.
package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"gopkg.in/yaml.v3"
)

// Schema defines validation rules for YAML.
type Schema struct {
	Required   []string            `yaml:"required"`
	Types      map[string]string   `yaml:"types"`
	Enums      map[string][]string `yaml:"enums"`
	MinLength  map[string]int      `yaml:"min_length"`
	MaxLength  map[string]int      `yaml:"max_length"`
	Patterns   map[string]string   `yaml:"patterns"`
	Nested     map[string]*Schema  `yaml:"nested"`
}

// Engine validates YAML against schemas.
type Engine struct{}

// NewEngine creates a new validation engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Validate validates YAML data against a schema.
func (e *Engine) Validate(data []byte, schema *Schema) (*models.LintReport, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	report := &models.LintReport{File: "yaml"}
	e.validateNode(&doc, schema, "", report)
	return report, nil
}

// ValidateFile validates a YAML file against a schema.
func (e *Engine) ValidateFile(path string, schema *Schema) (*models.LintReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	report := &models.LintReport{File: path}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		report.Issues = append(report.Issues, models.LintIssue{
			Line:     1,
			Severity: "error",
			Rule:     "parse-error",
			Message:  fmt.Sprintf("YAML parse error: %v", err),
		})
		return report, nil
	}

	e.validateNode(&doc, schema, "", report)
	return report, nil
}

// LoadSchema loads a schema from a YAML file.
func LoadSchema(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading schema file: %w", err)
	}

	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}

	return &schema, nil
}

// validateNode recursively validates a node against a schema.
func (e *Engine) validateNode(node *yaml.Node, schema *Schema, path string, report *models.LintReport) {
	if schema == nil || node == nil {
		return
	}

	// Handle DocumentNode — recurse into content
	if node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			e.validateNode(child, schema, path, report)
		}
		return
	}

	if node.Kind == yaml.MappingNode {
		e.validateMapping(node, schema, path, report)
	}
}

func (e *Engine) validateMapping(node *yaml.Node, schema *Schema, path string, report *models.LintReport) {
	// Check required fields
	seenKeys := make(map[string]bool)
	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i]
		seenKeys[key.Value] = true
	}

	for _, required := range schema.Required {
		if !seenKeys[required] {
			report.Issues = append(report.Issues, models.LintIssue{
				Line:     node.Line,
				Severity: "error",
				Rule:     "required-field",
				Message:  fmt.Sprintf("missing required field: %s", required),
			})
		}
	}

	// Validate each field
	for i := 0; i < len(node.Content)-1; i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		fieldPath := path + "." + key.Value
		if path == "" {
			fieldPath = key.Value
		}

		// Type checking
		if expectedType, ok := schema.Types[key.Value]; ok {
			actualType := getNodeType(val)
			if actualType != expectedType {
				report.Issues = append(report.Issues, models.LintIssue{
					Line:     val.Line,
					Severity: "error",
					Rule:     "type-check",
					Message:  fmt.Sprintf("field %q: expected %s, got %s", fieldPath, expectedType, actualType),
				})
			}
		}

		// Enum checking
		if values, ok := schema.Enums[key.Value]; ok {
			if val.Kind == yaml.ScalarNode {
				found := false
				for _, v := range values {
					if val.Value == v {
						found = true
						break
					}
				}
				if !found {
					report.Issues = append(report.Issues, models.LintIssue{
						Line:     val.Line,
						Severity: "error",
						Rule:     "enum-check",
						Message:  fmt.Sprintf("field %q: value %q not in allowed values: %v", fieldPath, val.Value, values),
					})
				}
			}
		}

		// Min length
		if minLen, ok := schema.MinLength[key.Value]; ok {
			if val.Kind == yaml.ScalarNode && len(val.Value) < minLen {
				report.Issues = append(report.Issues, models.LintIssue{
					Line:     val.Line,
					Severity: "error",
					Rule:     "min-length",
					Message:  fmt.Sprintf("field %q: length %d < minimum %d", fieldPath, len(val.Value), minLen),
				})
			}
		}

		// Max length
		if maxLen, ok := schema.MaxLength[key.Value]; ok {
			if val.Kind == yaml.ScalarNode && len(val.Value) > maxLen {
				report.Issues = append(report.Issues, models.LintIssue{
					Line:     val.Line,
					Severity: "error",
					Rule:     "max-length",
					Message:  fmt.Sprintf("field %q: length %d > maximum %d", fieldPath, len(val.Value), maxLen),
				})
			}
		}

		// Nested schema
		if nested, ok := schema.Nested[key.Value]; ok {
			e.validateNode(val, nested, fieldPath, report)
		}
	}
}

func getNodeType(node *yaml.Node) string {
	switch node.Kind {
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!null":
			return "null"
		case "!!bool":
			return "boolean"
		case "!!int":
			return "integer"
		case "!!float":
			return "number"
		default:
			return "string"
		}
	case yaml.MappingNode:
		return "object"
	case yaml.SequenceNode:
		return "array"
	default:
		return "unknown"
	}
}

// InferSchema infers a schema from YAML data.
func InferSchema(data []byte) (*Schema, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	schema := &Schema{
		Required:  make([]string, 0),
		Types:     make(map[string]string),
		Enums:     make(map[string][]string),
		Nested:    make(map[string]*Schema),
	}

	// Handle DocumentNode — unwrap to actual content
	root := &doc
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		root = doc.Content[0]
	}

	if root.Kind == yaml.MappingNode {
		for i := 0; i < len(root.Content)-1; i += 2 {
			key := root.Content[i]
			val := root.Content[i+1]

			schema.Required = append(schema.Required, key.Value)
			schema.Types[key.Value] = getNodeType(val)

			// Track enum values if scalar
			if val.Kind == yaml.ScalarNode && val.Value != "" {
				if _, ok := schema.Enums[key.Value]; !ok {
					schema.Enums[key.Value] = []string{}
				}
				schema.Enums[key.Value] = append(schema.Enums[key.Value], val.Value)
			}

			// Recurse into nested objects
			if val.Kind == yaml.MappingNode {
				nested, err := InferSchema(nodeToBytes(val))
				if err == nil {
					schema.Nested[key.Value] = nested
				}
			}
		}
	}

	// Deduplicate enum values
	for k, values := range schema.Enums {
		seen := make(map[string]bool)
		var unique []string
		for _, v := range values {
			if !seen[v] {
				seen[v] = true
				unique = append(unique, v)
			}
		}
		schema.Enums[k] = unique

		// If only one value, it's effectively a constant
		if len(unique) <= 1 {
			delete(schema.Enums, k)
		}
	}

	return schema, nil
}

func nodeToBytes(node *yaml.Node) []byte {
	var sb strings.Builder
	_ = yaml.NewEncoder(&sb).Encode(node)
	return []byte(sb.String())
}
