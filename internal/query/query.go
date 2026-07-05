// Package query provides YAMLPath expressions for querying YAML documents.
package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
)

// Engine evaluates YAMLPath expressions against YAML nodes.
type Engine struct{}

// NewEngine creates a new query engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Query evaluates a YAMLPath expression and returns matching nodes.
func (e *Engine) Query(root *models.Node, path string) ([]*models.Node, error) {
	segments, err := ParsePath(path)
	if err != nil {
		return nil, err
	}
	return e.resolve(root, segments)
}

// QueryFirst returns the first matching node.
func (e *Engine) QueryFirst(root *models.Node, path string) (*models.Node, error) {
	results, err := e.Query(root, path)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no match for path: %s", path)
	}
	return results[0], nil
}

// QueryValue returns the value of the first matching node.
func (e *Engine) QueryValue(root *models.Node, path string) (interface{}, error) {
	node, err := e.QueryFirst(root, path)
	if err != nil {
		return nil, err
	}
	return node.Value, nil
}

// ParsePath parses a YAMLPath string into segments.
func ParsePath(path string) ([]models.PathSegment, error) {
	if path == "" || path == "." {
		return nil, nil
	}

	path = strings.TrimPrefix(path, ".")
	var segments []models.PathSegment

	for path != "" {
		// Check for array index: [0], [1], etc.
		if strings.HasPrefix(path, "[") {
			end := strings.Index(path, "]")
			if end == -1 {
				return nil, fmt.Errorf("unclosed bracket in path: %s", path)
			}
			idxStr := path[1:end]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil, fmt.Errorf("invalid index: %s", idxStr)
			}
			segments = append(segments, models.PathSegment{
				Name:    fmt.Sprintf("[%d]", idx),
				Index:   idx,
				IsIndex: true,
			})
			path = path[end+1:]
			if strings.HasPrefix(path, ".") {
				path = path[1:]
			}
			continue
		}

		// Check for filter: [?@key==value]
		if strings.HasPrefix(path, "[?") {
			end := strings.Index(path, "]")
			if end == -1 {
				return nil, fmt.Errorf("unclosed filter in path: %s", path)
			}
			filterExpr := path[2:end]
			seg, err := parseFilter(filterExpr)
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
			path = path[end+1:]
			if strings.HasPrefix(path, ".") {
				path = path[1:]
			}
			continue
		}

		// Regular key segment
		nextDot := strings.Index(path, ".")
		nextBracket := strings.Index(path, "[")
		var segName string

		if nextDot == -1 && nextBracket == -1 {
			segName = path
			path = ""
		} else if nextDot == -1 {
			segName = path[:nextBracket]
			path = path[nextBracket:]
		} else if nextBracket == -1 {
			segName = path[:nextDot]
			path = path[nextDot+1:]
		} else {
			if nextDot < nextBracket {
				segName = path[:nextDot]
				path = path[nextDot+1:]
			} else {
				segName = path[:nextBracket]
				path = path[nextBracket:]
			}
		}

		segments = append(segments, models.PathSegment{
			Name: segName,
		})
	}

	return segments, nil
}

// parseFilter parses a filter expression like ?@key==value.
func parseFilter(expr string) (models.PathSegment, error) {
	seg := models.PathSegment{IsFilter: true}

	// Support: ?@key==value, ?@key!=value, ?@key>value, ?@key<value
	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if idx := strings.Index(expr, op); idx != -1 {
			seg.Name = expr[:idx]
			seg.FilterOp = op
			seg.FilterVal = expr[idx+len(op):]
			return seg, nil
		}
	}

	return seg, fmt.Errorf("invalid filter expression: %s", expr)
}

// resolve resolves path segments against a root node.
func (e *Engine) resolve(node *models.Node, segments []models.PathSegment) ([]*models.Node, error) {
	if len(segments) == 0 {
		return []*models.Node{node}, nil
	}

	seg := segments[0]
	rest := segments[1:]

	var results []*models.Node

	if seg.IsFilter {
		// Filter children
		if node.Type == models.NodeMapping || node.Type == models.NodeSequence {
			for _, child := range node.Children {
				matches, err := matchFilter(child, seg)
				if err != nil {
					return nil, err
				}
				if matches {
					r, err := e.resolve(child, rest)
					if err != nil {
						return nil, err
					}
					results = append(results, r...)
				}
			}
		}
		return results, nil
	}

	if seg.IsIndex {
		if node.Type == models.NodeSequence {
			if seg.Index < len(node.Children) {
				return e.resolve(node.Children[seg.Index], rest)
			}
			return nil, fmt.Errorf("index %d out of range (len=%d)", seg.Index, len(node.Children))
		}
		return nil, fmt.Errorf("cannot index into non-sequence node")
	}

	// Named key or wildcard
	if node.Type == models.NodeMapping || node.Type == models.NodeSequence {
		if seg.Name == "*" {
			// Wildcard: match any child
			for _, child := range node.Children {
				r, err := e.resolve(child, rest)
				if err != nil {
					continue // skip non-matching
				}
				results = append(results, r...)
			}
			return results, nil
		}

		if node.Type == models.NodeMapping {
			child := node.GetChild(seg.Name)
			if child == nil {
				return nil, fmt.Errorf("key not found: %s", seg.Name)
			}
			return e.resolve(child, rest)
		}
	}

	return nil, fmt.Errorf("cannot traverse into %s node with key %s", node.Type, seg.Name)
}

// matchFilter checks if a node matches a filter segment.
func matchFilter(node *models.Node, seg models.PathSegment) (bool, error) {
	child := node.GetChild(seg.Name)
	if child == nil {
		return false, nil
	}

	valStr := fmt.Sprintf("%v", child.Value)

	switch seg.FilterOp {
	case "==":
		return valStr == seg.FilterVal, nil
	case "!=":
		return valStr != seg.FilterVal, nil
	case ">":
		return compareValues(valStr, seg.FilterVal, func(a, b float64) bool { return a > b })
	case "<":
		return compareValues(valStr, seg.FilterVal, func(a, b float64) bool { return a < b })
	case ">=":
		return compareValues(valStr, seg.FilterVal, func(a, b float64) bool { return a >= b })
	case "<=":
		return compareValues(valStr, seg.FilterVal, func(a, b float64) bool { return a <= b })
	default:
		return false, fmt.Errorf("unknown filter operator: %s", seg.FilterOp)
	}
}

// compareValues compares two numeric string values.
func compareValues(a, b string, cmp func(float64, float64) bool) (bool, error) {
	af, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return false, fmt.Errorf("cannot compare non-numeric values: %s", a)
	}
	bf, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return false, fmt.Errorf("cannot compare non-numeric values: %s", b)
	}
	return cmp(af, bf), nil
}

// SetValue sets a value at the given path in the YAML tree.
func (e *Engine) SetValue(root *models.Node, path string, value interface{}) error {
	segments, err := ParsePath(path)
	if err != nil {
		return err
	}
	return e.setValue(root, segments, value)
}

func (e *Engine) setValue(node *models.Node, segments []models.PathSegment, value interface{}) error {
	if len(segments) == 0 {
		return fmt.Errorf("empty path")
	}

	seg := segments[0]
	rest := segments[1:]

	if len(rest) == 0 {
		// Set value here
		if node.Type == models.NodeMapping {
			child := node.GetChild(seg.Name)
			if child != nil {
				child.Type = models.NodeScalar
				child.Value = value
				child.Children = nil
			} else {
				node.AddChild(models.NewScalar(seg.Name, value))
			}
			return nil
		}
		return fmt.Errorf("cannot set value on %s node", node.Type)
	}

	if node.Type == models.NodeMapping {
		child := node.GetChild(seg.Name)
		if child == nil {
			child = models.NewMapping(seg.Name)
			node.AddChild(child)
		}
		return e.setValue(child, rest, value)
	}

	return fmt.Errorf("cannot traverse into %s node", node.Type)
}

// DeleteValue deletes a value at the given path.
func (e *Engine) DeleteValue(root *models.Node, path string) error {
	segments, err := ParsePath(path)
	if err != nil {
		return err
	}

	if len(segments) == 0 {
		return fmt.Errorf("empty path")
	}

	if len(segments) == 1 && root.Type == models.NodeMapping {
		for i, child := range root.Children {
			if child.Key == segments[0].Name {
				root.Children = append(root.Children[:i], root.Children[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("key not found: %s", segments[0].Name)
	}

	return e.traverseAndDelete(root, segments)
}

func (e *Engine) traverseAndDelete(node *models.Node, segments []models.PathSegment) error {
	if len(segments) <= 1 {
		return nil
	}

	seg := segments[0]
	rest := segments[1:]

	if node.Type == models.NodeMapping {
		child := node.GetChild(seg.Name)
		if child == nil {
			return fmt.Errorf("key not found: %s", seg.Name)
		}

		if len(rest) == 1 {
			// Delete the child
			for i, c := range node.Children {
				if c.Key == rest[0].Name {
					node.Children = append(node.Children[:i], node.Children[i+1:]...)
					return nil
				}
			}
			return fmt.Errorf("key not found: %s", rest[0].Name)
		}

		return e.traverseAndDelete(child, rest)
	}

	return fmt.Errorf("cannot traverse into %s node", node.Type)
}
