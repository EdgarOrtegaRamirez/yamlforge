// Package format provides YAML formatting and pretty-printing.
package format

import (
	"fmt"
	"strings"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"gopkg.in/yaml.v3"
)

// Options controls formatting behavior.
type Options struct {
	Indent      int
	SortKeys    bool
	FlowStyle   bool
	Width       int
	Prefix      string
}

// DefaultOptions returns default formatting options.
func DefaultOptions() Options {
	return Options{
		Indent:   2,
		SortKeys: false,
		FlowStyle: false,
		Width:    80,
	}
}

// Formatter formats YAML content.
type Formatter struct {
	opts Options
}

// NewFormatter creates a new Formatter.
func NewFormatter(opts Options) *Formatter {
	return &Formatter{opts: opts}
}

// Format formats YAML bytes.
func (f *Formatter) Format(data []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if f.opts.SortKeys {
		sortYAMLNode(&doc)
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(f.opts.Indent)

	// Ensure we encode a document node
	if doc.Kind != yaml.DocumentNode {
		doc = yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{&doc},
		}
	}

	if err := encoder.Encode(&doc); err != nil {
		return nil, fmt.Errorf("encoding YAML: %w", err)
	}
	encoder.Close()

	return []byte(buf.String()), nil
}

// FormatNode formats a YAML node tree back to bytes.
func (f *Formatter) FormatNode(node *models.Node) ([]byte, error) {
	yn := nodeToYAML(node)

	if f.opts.SortKeys {
		sortYAMLNode(yn)
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(f.opts.Indent)

	// Ensure we encode a document node
	if yn.Kind != yaml.DocumentNode {
		yn = &yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{yn},
		}
	}

	if err := encoder.Encode(yn); err != nil {
		return nil, fmt.Errorf("encoding YAML: %w", err)
	}
	encoder.Close()

	return []byte(buf.String()), nil
}

// Minify removes unnecessary whitespace from YAML.
func Minify(data []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if doc.Kind != yaml.DocumentNode {
		doc = yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{&doc},
		}
	}

	if err := encoder.Encode(&doc); err != nil {
		return nil, fmt.Errorf("encoding YAML: %w", err)
	}
	encoder.Close()

	return []byte(buf.String()), nil
}

// PrettyPrint formats YAML with nice indentation.
func PrettyPrint(data []byte, indent int) ([]byte, error) {
	f := NewFormatter(Options{Indent: indent})
	return f.Format(data)
}

// SortKeys sorts all mapping keys alphabetically.
func SortKeys(data []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	sortYAMLNode(&doc)

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if doc.Kind != yaml.DocumentNode {
		doc = yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{&doc},
		}
	}

	if err := encoder.Encode(&doc); err != nil {
		return nil, fmt.Errorf("encoding YAML: %w", err)
	}
	encoder.Close()

	return []byte(buf.String()), nil
}

// sortYAMLNode recursively sorts mapping keys.
func sortYAMLNode(node *yaml.Node) {
	if node == nil {
		return
	}

	if node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			sortYAMLNode(child)
		}
		return
	}

	if node.Kind == yaml.MappingNode {
		// Sort key-value pairs by key
		for i := 0; i < len(node.Content)-1; i += 2 {
			for j := i + 2; j < len(node.Content); j += 2 {
				if node.Content[j].Value < node.Content[i].Value {
					// Swap key and value
					node.Content[i], node.Content[j] = node.Content[j], node.Content[i]
					node.Content[i+1], node.Content[j+1] = node.Content[j+1], node.Content[i+1]
				}
			}
		}

		// Recursively sort children
		for _, child := range node.Content {
			sortYAMLNode(child)
		}
	} else if node.Kind == yaml.SequenceNode {
		for _, child := range node.Content {
			sortYAMLNode(child)
		}
	}
}

// nodeToYAML converts our internal node to a yaml.v3 node.
func nodeToYAML(node *models.Node) *yaml.Node {
	if node == nil {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"}
	}

	yn := &yaml.Node{
		Kind:  yamlToKind(node.Type),
		Value: node.Key,
		Tag:   node.Tag,
	}

	if node.Comment != "" {
		yn.HeadComment = node.Comment
	}

	switch node.Type {
	case models.NodeScalar:
		yn.Kind = yaml.ScalarNode
		if node.Value == nil {
			yn.Tag = "!!null"
			yn.Value = ""
		} else {
			yn.Value = fmt.Sprintf("%v", node.Value)
			switch node.Value.(type) {
			case bool:
				yn.Tag = "!!bool"
			case int:
				yn.Tag = "!!int"
			case float64:
				yn.Tag = "!!float"
			default:
				yn.Tag = "!!str"
			}
		}

	case models.NodeMapping:
		yn.Kind = yaml.MappingNode
		for _, child := range node.Children {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: child.Key,
				Tag:   "!!str",
			}
			valNode := nodeToYAML(child)
			yn.Content = append(yn.Content, keyNode, valNode)
		}

	case models.NodeSequence:
		yn.Kind = yaml.SequenceNode
		for _, child := range node.Children {
			yn.Content = append(yn.Content, nodeToYAML(child))
		}
	}

	return yn
}

func yamlToKind(nt models.NodeType) yaml.Kind {
	switch nt {
	case models.NodeScalar:
		return yaml.ScalarNode
	case models.NodeMapping:
		return yaml.MappingNode
	case models.NodeSequence:
		return yaml.SequenceNode
	default:
		return yaml.ScalarNode
	}
}
