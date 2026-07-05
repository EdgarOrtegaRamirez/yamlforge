// Package merge provides deep merging of YAML node trees.
package merge

import (
	"fmt"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
)

// Engine merges YAML node trees.
type Engine struct {
	strategy models.MergeStrategy
}

// NewEngine creates a new merge engine with the given strategy.
func NewEngine(strategy models.MergeStrategy) *Engine {
	return &Engine{strategy: strategy}
}

// Merge merges source into target and returns the result.
func (e *Engine) Merge(target, source *models.Node) (*models.Node, error) {
	if target == nil {
		return e.cloneNode(source), nil
	}
	if source == nil {
		return e.cloneNode(target), nil
	}

	if target.Type != source.Type {
		switch e.strategy {
		case models.MergeReplace:
			return e.cloneNode(source), nil
		case models.MergeKeep:
			return e.cloneNode(target), nil
		default:
			return nil, fmt.Errorf("cannot merge %s with %s", target.Type, source.Type)
		}
	}

	switch target.Type {
	case models.NodeMapping:
		return e.mergeMappings(target, source)
	case models.NodeSequence:
		return e.mergeSequences(target, source)
	case models.NodeScalar:
		return e.mergeScalars(target, source)
	default:
		return e.cloneNode(target), nil
	}
}

// mergeMappings merges two mapping nodes.
func (e *Engine) mergeMappings(target, source *models.Node) (*models.Node, error) {
	result := &models.Node{
		Type:     models.NodeMapping,
		Key:      target.Key,
		Children: make([]*models.Node, 0),
	}

	// Clone target children first
	seenKeys := make(map[string]bool)
	for _, child := range target.Children {
		cloned := e.cloneNode(child)
		result.Children = append(result.Children, cloned)
		seenKeys[child.Key] = true
	}

	// Merge source children
	for _, srcChild := range source.Children {
		if existing := result.GetChild(srcChild.Key); existing != nil {
			// Key exists in both - merge recursively
			merged, err := e.Merge(existing, srcChild)
			if err != nil {
				return nil, err
			}
			// Replace existing with merged
			for i, child := range result.Children {
				if child.Key == srcChild.Key {
					result.Children[i] = merged
					break
				}
			}
		} else {
			// New key from source
			result.Children = append(result.Children, e.cloneNode(srcChild))
		}
		seenKeys[srcChild.Key] = true
	}

	return result, nil
}

// mergeSequences merges two sequence nodes.
func (e *Engine) mergeSequences(target, source *models.Node) (*models.Node, error) {
	result := &models.Node{
		Type:     models.NodeSequence,
		Key:      target.Key,
		Children: make([]*models.Node, 0),
	}

	switch e.strategy {
	case models.MergeAppend:
		// Append all items
		for _, child := range target.Children {
			result.Children = append(result.Children, e.cloneNode(child))
		}
		for _, child := range source.Children {
			result.Children = append(result.Children, e.cloneNode(child))
		}

	case models.MergeReplace:
		// Replace with source
		for _, child := range source.Children {
			result.Children = append(result.Children, e.cloneNode(child))
		}

	case models.MergeDeep:
		// Deep merge by index
		maxLen := len(target.Children)
		if len(source.Children) > maxLen {
			maxLen = len(source.Children)
		}
		for i := 0; i < maxLen; i++ {
			if i >= len(target.Children) {
				result.Children = append(result.Children, e.cloneNode(source.Children[i]))
			} else if i >= len(source.Children) {
				result.Children = append(result.Children, e.cloneNode(target.Children[i]))
			} else {
				merged, err := e.Merge(target.Children[i], source.Children[i])
				if err != nil {
					return nil, err
				}
				result.Children = append(result.Children, merged)
			}
		}

	case models.MergeKeep:
		// Keep target items
		for _, child := range target.Children {
			result.Children = append(result.Children, e.cloneNode(child))
		}
	}

	return result, nil
}

// mergeScalars merges two scalar nodes.
func (e *Engine) mergeScalars(target, source *models.Node) (*models.Node, error) {
	switch e.strategy {
	case models.MergeReplace:
		return e.cloneNode(source), nil
	case models.MergeKeep:
		return e.cloneNode(target), nil
	default:
		return e.cloneNode(source), nil
	}
}

// cloneNode creates a deep copy of a node.
func (e *Engine) cloneNode(n *models.Node) *models.Node {
	if n == nil {
		return nil
	}

	clone := &models.Node{
		Type:     n.Type,
		Key:      n.Key,
		Value:    n.Value,
		Comment:  n.Comment,
		Line:     n.Line,
		Column:   n.Column,
		Tag:      n.Tag,
		IsAnchor: n.IsAnchor,
		Anchor:   n.Anchor,
	}

	if len(n.Children) > 0 {
		clone.Children = make([]*models.Node, len(n.Children))
		for i, child := range n.Children {
			clone.Children[i] = e.cloneNode(child)
		}
	}

	return clone
}

// MergeFiles merges multiple YAML node trees.
func MergeFiles(strategy models.MergeStrategy, nodes ...*models.Node) (*models.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes to merge")
	}

	engine := NewEngine(strategy)
	result := nodes[0]

	for i := 1; i < len(nodes); i++ {
		var err error
		result, err = engine.Merge(result, nodes[i])
		if err != nil {
			return nil, fmt.Errorf("merging document %d: %w", i, err)
		}
	}

	return result, nil
}
