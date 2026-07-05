// Package filter provides multi-document YAML filtering.
package filter

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Engine filters multi-document YAML files.
type Engine struct{}

// NewEngine creates a new filter engine.
func NewEngine() *Engine {
	return &Engine{}
}

// FilterByPath filters documents by YAMLPath existence.
func (e *Engine) FilterByPath(data []byte, path string) ([]byte, error) {
	docs, err := e.splitDocuments(data)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, doc := range docs {
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(doc), &node); err != nil {
			continue
		}

		if hasPath(&node, path) {
			result = append(result, doc)
		}
	}

	return []byte(strings.Join(result, "---\n")), nil
}

// FilterByContent filters documents containing a string.
func (e *Engine) FilterByContent(data []byte, search string) ([]byte, error) {
	docs, err := e.splitDocuments(data)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, doc := range docs {
		if strings.Contains(doc, search) {
			result = append(result, doc)
		}
	}

	return []byte(strings.Join(result, "---\n")), nil
}

// FilterByIndex returns a specific document by index (0-based).
func (e *Engine) FilterByIndex(data []byte, index int) ([]byte, error) {
	docs, err := e.splitDocuments(data)
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= len(docs) {
		return nil, fmt.Errorf("document index %d out of range (0-%d)", index, len(docs)-1)
	}

	return []byte(docs[index]), nil
}

// FilterByRange returns a range of documents.
func (e *Engine) FilterByRange(data []byte, start, end int) ([]byte, error) {
	docs, err := e.splitDocuments(data)
	if err != nil {
		return nil, err
	}

	if start < 0 {
		start = 0
	}
	if end >= len(docs) {
		end = len(docs) - 1
	}
	if start > end {
		return nil, fmt.Errorf("invalid range: %d-%d", start, end)
	}

	var result []string
	for i := start; i <= end; i++ {
		result = append(result, docs[i])
	}

	return []byte(strings.Join(result, "---\n")), nil
}

// CountDocuments counts the number of YAML documents.
func (e *Engine) CountDocuments(data []byte) (int, error) {
	docs, err := e.splitDocuments(data)
	if err != nil {
		return 0, err
	}
	return len(docs), nil
}

// SplitDocuments splits multi-document YAML into individual documents.
func (e *Engine) SplitDocuments(data []byte) ([]string, error) {
	return e.splitDocuments(data)
}

// splitDocuments splits YAML content on document separators.
func (e *Engine) splitDocuments(data []byte) ([]string, error) {
	content := string(data)
	var docs []string

	// Split on --- separator
	parts := strings.Split(content, "---")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			docs = append(docs, trimmed)
		}
	}

	return docs, nil
}

// hasPath checks if a YAML node has a given path.
func hasPath(node *yaml.Node, path string) bool {
	if node == nil {
		return false
	}

	segments := strings.Split(strings.TrimPrefix(path, "."), ".")
	return traversePath(node, segments)
}

func traversePath(node *yaml.Node, segments []string) bool {
	if len(segments) == 0 {
		return true
	}

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content)-1; i += 2 {
			key := node.Content[i]
			val := node.Content[i+1]
			if key.Value == segments[0] {
				return traversePath(val, segments[1:])
			}
		}
	}

	return false
}

// FilterFile filters a YAML file.
func (e *Engine) FilterFile(path string, filterType, value string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	switch filterType {
	case "path":
		return e.FilterByPath(data, value)
	case "content":
		return e.FilterByContent(data, value)
	case "index":
		var idx int
		if _, err := fmt.Sscanf(value, "%d", &idx); err != nil {
			return nil, fmt.Errorf("invalid index: %s", value)
		}
		return e.FilterByIndex(data, idx)
	default:
		return nil, fmt.Errorf("unknown filter type: %s", filterType)
	}
}
