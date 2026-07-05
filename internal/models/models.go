// Package models provides core data types for YAML Forge.
package models

import (
	"fmt"
	"strings"
)

// NodeType represents the type of a YAML node.
type NodeType int

const (
	NodeScalar NodeType = iota
	NodeMapping
	NodeSequence
	NodeAlias
	NodeNull
)

func (n NodeType) String() string {
	switch n {
	case NodeScalar:
		return "scalar"
	case NodeMapping:
		return "mapping"
	case NodeSequence:
		return "sequence"
	case NodeAlias:
		return "alias"
	case NodeNull:
		return "null"
	default:
		return "unknown"
	}
}

// Node represents a YAML document node.
type Node struct {
	Type     NodeType
	Key      string
	Value    interface{}
	Children []*Node
	Comment  string
	Line     int
	Column   int
	Tag      string
	IsAnchor bool
	Anchor   string
}

// NewScalar creates a scalar node.
func NewScalar(key string, value interface{}) *Node {
	return &Node{
		Type:  NodeScalar,
		Key:   key,
		Value: value,
	}
}

// NewMapping creates a mapping node.
func NewMapping(key string) *Node {
	return &Node{
		Type:     NodeMapping,
		Key:      key,
		Children: make([]*Node, 0),
	}
}

// NewSequence creates a sequence node.
func NewSequence(key string) *Node {
	return &Node{
		Type:     NodeSequence,
		Key:      key,
		Children: make([]*Node, 0),
	}
}

// AddChild adds a child node.
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}

// GetChild finds a child by key.
func (n *Node) GetChild(key string) *Node {
	for _, child := range n.Children {
		if child.Key == key {
			return child
		}
	}
	return nil
}

// String returns a string representation.
func (n *Node) String() string {
	if n == nil {
		return "<nil>"
	}
	switch n.Type {
	case NodeScalar:
		return fmt.Sprintf("%s: %v", n.Key, n.Value)
	case NodeMapping:
		return fmt.Sprintf("%s: {%d children}", n.Key, len(n.Children))
	case NodeSequence:
		return fmt.Sprintf("%s: [%d items]", n.Key, len(n.Children))
	default:
		return fmt.Sprintf("%s: %s", n.Key, n.Type)
	}
}

// Path represents a YAMLPath expression.
type Path struct {
	Segments []PathSegment
}

// PathSegment is a single segment of a YAMLPath.
type PathSegment struct {
	Name      string
	Index     int
	IsIndex   bool
	IsFilter  bool
	FilterOp  string
	FilterVal string
}

// String returns the string representation of a path.
func (p Path) String() string {
	parts := make([]string, len(p.Segments))
	for i, seg := range p.Segments {
		if seg.IsIndex {
			parts[i] = fmt.Sprintf("[%d]", seg.Index)
		} else {
			parts[i] = seg.Name
		}
	}
	return strings.Join(parts, ".")
}

// DiffOp represents a diff operation.
type DiffOp int

const (
	DiffAdd DiffOp = iota
	DiffRemove
	DiffModify
	DiffEqual
)

func (d DiffOp) String() string {
	switch d {
	case DiffAdd:
		return "+"
	case DiffRemove:
		return "-"
	case DiffModify:
		return "~"
	case DiffEqual:
		return "="
	default:
		return "?"
	}
}

// DiffEntry represents a single diff entry.
type DiffEntry struct {
	Op       DiffOp
	Path     string
	OldValue interface{}
	NewValue interface{}
	Type     string
}

// LintIssue represents a linting issue.
type LintIssue struct {
	Line     int
	Column   int
	Severity string // error, warning, info
	Rule     string
	Message  string
}

// LintReport contains all lint issues for a file.
type LintReport struct {
	File  string
	Issues []LintIssue
}

// HasErrors returns true if there are any error-level issues.
func (r *LintReport) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}

// ErrorCount returns the number of error-level issues.
func (r *LintReport) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warning-level issues.
func (r *LintReport) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == "warning" {
			count++
		}
	}
	return count
}

// MergeStrategy defines how conflicts are resolved during merge.
type MergeStrategy int

const (
	MergeReplace MergeStrategy = iota
	MergeAppend
	MergeDeep
	MergeKeep
)

// YAMLStats contains statistics about a YAML document.
type YAMLStats struct {
	TotalNodes    int
	MappingNodes  int
	SequenceNodes int
	ScalarNodes   int
	MaxDepth      int
	TotalKeys     int
	UniqueKeys    int
	KeyFrequency  map[string]int
	LineCount     int
	AnchorCount   int
	AliasCount    int
	TagCount      int
}
