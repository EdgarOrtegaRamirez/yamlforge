// Package diff provides semantic YAML diffing.
package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
)

// Engine compares two YAML node trees and produces diff entries.
type Engine struct{}

// NewEngine creates a new diff engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Diff compares two YAML trees and returns diff entries.
func (e *Engine) Diff(a, b *models.Node) []models.DiffEntry {
	var entries []models.DiffEntry
	e.diffNodes(a, b, "", &entries)
	return entries
}

// diffNodes recursively compares two nodes.
func (e *Engine) diffNodes(a, b *models.Node, path string, entries *[]models.DiffEntry) {
	if a == nil && b == nil {
		return
	}

	if a == nil {
		*entries = append(*entries, models.DiffEntry{
			Op:       models.DiffAdd,
			Path:     path,
			NewValue: nodeToInterface(b),
			Type:     b.Type.String(),
		})
		return
	}

	if b == nil {
		*entries = append(*entries, models.DiffEntry{
			Op:       models.DiffRemove,
			Path:     path,
			OldValue: nodeToInterface(a),
			Type:     a.Type.String(),
		})
		return
	}

	// Different types
	if a.Type != b.Type {
		*entries = append(*entries, models.DiffEntry{
			Op:       models.DiffModify,
			Path:     path,
			OldValue: nodeToInterface(a),
			NewValue: nodeToInterface(b),
			Type:     fmt.Sprintf("%s -> %s", a.Type, b.Type),
		})
		return
	}

	switch a.Type {
	case models.NodeScalar:
		if !valuesEqual(a.Value, b.Value) {
			*entries = append(*entries, models.DiffEntry{
				Op:       models.DiffModify,
				Path:     path,
				OldValue: a.Value,
				NewValue: b.Value,
				Type:     "scalar",
			})
		}

	case models.NodeMapping:
		e.diffMappings(a, b, path, entries)

	case models.NodeSequence:
		e.diffSequences(a, b, path, entries)
	}
}

// diffMappings compares two mapping nodes.
func (e *Engine) diffMappings(a, b *models.Node, path string, entries *[]models.DiffEntry) {
	// Build maps of children
	aMap := make(map[string]*models.Node)
	for _, child := range a.Children {
		aMap[child.Key] = child
	}
	bMap := make(map[string]*models.Node)
	for _, child := range b.Children {
		bMap[child.Key] = child
	}

	// Get all keys in sorted order
	allKeys := make(map[string]bool)
	for k := range aMap {
		allKeys[k] = true
	}
	for k := range bMap {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		childPath := path + "." + key
		if path == "" {
			childPath = key
		}

		aChild, aExists := aMap[key]
		bChild, bExists := bMap[key]

		if !aExists {
			*entries = append(*entries, models.DiffEntry{
				Op:       models.DiffAdd,
				Path:     childPath,
				NewValue: nodeToInterface(bChild),
				Type:     bChild.Type.String(),
			})
		} else if !bExists {
			*entries = append(*entries, models.DiffEntry{
				Op:       models.DiffRemove,
				Path:     childPath,
				OldValue: nodeToInterface(aChild),
				Type:     aChild.Type.String(),
			})
		} else {
			e.diffNodes(aChild, bChild, childPath, entries)
		}
	}
}

// diffSequences compares two sequence nodes.
func (e *Engine) diffSequences(a, b *models.Node, path string, entries *[]models.DiffEntry) {
	maxLen := len(a.Children)
	if len(b.Children) > maxLen {
		maxLen = len(b.Children)
	}

	for i := 0; i < maxLen; i++ {
		itemPath := fmt.Sprintf("%s[%d]", path, i)

		if i >= len(a.Children) {
			*entries = append(*entries, models.DiffEntry{
				Op:       models.DiffAdd,
				Path:     itemPath,
				NewValue: nodeToInterface(b.Children[i]),
				Type:     b.Children[i].Type.String(),
			})
		} else if i >= len(b.Children) {
			*entries = append(*entries, models.DiffEntry{
				Op:       models.DiffRemove,
				Path:     itemPath,
				OldValue: nodeToInterface(a.Children[i]),
				Type:     a.Children[i].Type.String(),
			})
		} else {
			e.diffNodes(a.Children[i], b.Children[i], itemPath, entries)
		}
	}
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// nodeToInterface converts a node to an interface{} for display.
func nodeToInterface(n *models.Node) interface{} {
	if n == nil {
		return nil
	}
	switch n.Type {
	case models.NodeScalar:
		return n.Value
	case models.NodeMapping:
		m := make(map[string]interface{})
		for _, child := range n.Children {
			m[child.Key] = nodeToInterface(child)
		}
		return m
	case models.NodeSequence:
		var arr []interface{}
		for _, child := range n.Children {
			arr = append(arr, nodeToInterface(child))
		}
		return arr
	default:
		return nil
	}
}

// FormatDiff formats diff entries as a human-readable string.
func FormatDiff(entries []models.DiffEntry, format string) string {
	switch format {
	case "json":
		return formatDiffJSON(entries)
	case "compact":
		return formatDiffCompact(entries)
	default:
		return formatDiffText(entries)
	}
}

// formatDiffText formats diff entries in unified text format.
func formatDiffText(entries []models.DiffEntry) string {
	var sb strings.Builder
	for _, entry := range entries {
		switch entry.Op {
		case models.DiffAdd:
			sb.WriteString(fmt.Sprintf("+ %s: %v\n", entry.Path, entry.NewValue))
		case models.DiffRemove:
			sb.WriteString(fmt.Sprintf("- %s: %v\n", entry.Path, entry.OldValue))
		case models.DiffModify:
			sb.WriteString(fmt.Sprintf("~ %s: %v -> %v\n", entry.Path, entry.OldValue, entry.NewValue))
		}
	}
	return sb.String()
}

// formatDiffCompact formats diff entries compactly.
func formatDiffCompact(entries []models.DiffEntry) string {
	var sb strings.Builder
	for _, entry := range entries {
		sb.WriteString(fmt.Sprintf("%s %s\n", entry.Op, entry.Path))
	}
	return sb.String()
}

// formatDiffJSON formats diff entries as JSON-like output.
func formatDiffJSON(entries []models.DiffEntry) string {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, entry := range entries {
		sb.WriteString(fmt.Sprintf(`  {"op": "%s", "path": "%s"`, entry.Op, entry.Path))
		if entry.OldValue != nil {
			sb.WriteString(fmt.Sprintf(`, "old": "%v"`, entry.OldValue))
		}
		if entry.NewValue != nil {
			sb.WriteString(fmt.Sprintf(`, "new": "%v"`, entry.NewValue))
		}
		sb.WriteString("}")
		if i < len(entries)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("]")
	return sb.String()
}

// HasChanges returns true if there are any non-equal diff entries.
func HasChanges(entries []models.DiffEntry) bool {
	for _, e := range entries {
		if e.Op != models.DiffEqual {
			return true
		}
	}
	return false
}

// Summary returns a summary of the diff.
func Summary(entries []models.DiffEntry) (added, removed, modified int) {
	for _, e := range entries {
		switch e.Op {
		case models.DiffAdd:
			added++
		case models.DiffRemove:
			removed++
		case models.DiffModify:
			modified++
		}
	}
	return
}
