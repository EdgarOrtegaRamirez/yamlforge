// Package parser provides YAML parsing with comment preservation and anchor/alias support.
package parser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"gopkg.in/yaml.v3"
)

// Parser parses YAML content into our Node tree.
type Parser struct {
	anchors map[string]*models.Node
}

// New creates a new Parser.
func New() *Parser {
	return &Parser{
		anchors: make(map[string]*models.Node),
	}
}

// ParseFile parses a YAML file.
func (p *Parser) ParseFile(path string) ([]*models.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return p.ParseBytes(data)
}

// ParseBytes parses YAML bytes into a slice of document nodes.
func (p *Parser) ParseBytes(data []byte) ([]*models.Node, error) {
	p.anchors = make(map[string]*models.Node)

	var docs []*yaml.Node
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var doc yaml.Node
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decoding YAML: %w", err)
		}
		docs = append(docs, &doc)
	}

	result := make([]*models.Node, len(docs))
	for i, doc := range docs {
		node, err := p.convertNode(doc)
		if err != nil {
			return nil, fmt.Errorf("converting document %d: %w", i, err)
		}
		result[i] = node
	}
	return result, nil
}

// ParseString parses a YAML string.
func (p *Parser) ParseString(s string) ([]*models.Node, error) {
	return p.ParseBytes([]byte(s))
}

// ParseStream parses multiple YAML documents from a reader.
func (p *Parser) ParseStream(reader io.Reader) ([]*models.Node, error) {
	p.anchors = make(map[string]*models.Node)

	var docs []*yaml.Node
	decoder := yaml.NewDecoder(reader)
	for {
		var doc yaml.Node
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decoding YAML stream: %w", err)
		}
		docs = append(docs, &doc)
	}

	result := make([]*models.Node, len(docs))
	for i, doc := range docs {
		node, err := p.convertNode(doc)
		if err != nil {
			return nil, fmt.Errorf("converting document %d: %w", i, err)
		}
		result[i] = node
	}
	return result, nil
}

// convertNode converts a yaml.v3 Node to our internal Node.
func (p *Parser) convertNode(yn *yaml.Node) (*models.Node, error) {
	if yn == nil {
		return models.NewScalar("", nil), nil
	}

	node := &models.Node{
		Line:   yn.Line,
		Column: yn.Column,
	}

	// Extract comments
	if yn.HeadComment != "" {
		node.Comment = yn.HeadComment
	}
	if yn.LineComment != "" {
		if node.Comment != "" {
			node.Comment += " " + yn.LineComment
		} else {
			node.Comment = yn.LineComment
		}
	}
	if yn.FootComment != "" {
		if node.Comment != "" {
			node.Comment += " " + yn.FootComment
		} else {
			node.Comment = yn.FootComment
		}
	}

	// Handle tags
	if yn.Tag != "" {
		node.Tag = yn.Tag
	}

	// Handle anchors
	if yn.Anchor != "" {
		node.IsAnchor = true
		node.Anchor = yn.Anchor
	}

	switch yn.Kind {
	case yaml.DocumentNode:
		if len(yn.Content) > 0 {
			return p.convertNode(yn.Content[0])
		}
		return node, nil

	case yaml.MappingNode:
		node.Type = models.NodeMapping
		node.Key = yn.Value
		for i := 0; i < len(yn.Content); i += 2 {
			if i+1 >= len(yn.Content) {
				break
			}
			keyNode := yn.Content[i]
			valNode := yn.Content[i+1]

			child, err := p.convertNode(valNode)
			if err != nil {
				return nil, err
			}
			child.Key = keyNode.Value
			node.AddChild(child)
		}
		// Register anchor
		if node.IsAnchor {
			p.anchors[node.Anchor] = node
		}
		return node, nil

	case yaml.SequenceNode:
		node.Type = models.NodeSequence
		node.Key = yn.Value
		for _, content := range yn.Content {
			child, err := p.convertNode(content)
			if err != nil {
				return nil, err
			}
			node.AddChild(child)
		}
		if node.IsAnchor {
			p.anchors[node.Anchor] = node
		}
		return node, nil

	case yaml.ScalarNode:
		node.Type = models.NodeScalar
		node.Key = yn.Value
		node.Value = p.parseScalarValue(yn)
		if node.IsAnchor {
			p.anchors[node.Anchor] = node
		}
		return node, nil

	case yaml.AliasNode:
		node.Type = models.NodeAlias
		if yn.Anchor != "" {
			if anchor, ok := p.anchors[yn.Anchor]; ok {
				return anchor, nil
			}
		}
		return node, nil
	}

	return node, nil
}

// parseScalarValue parses a scalar value into its Go type.
func (p *Parser) parseScalarValue(yn *yaml.Node) interface{} {
	if yn.Value == "" && yn.Tag == "!!null" {
		return nil
	}

	switch yn.Tag {
	case "!!null":
		return nil
	case "!!bool":
		return yn.Value == "true" || yn.Value == "True" || yn.Value == "TRUE" || yn.Value == "yes" || yn.Value == "Yes"
	case "!!int":
		var v int
		if _, err := fmt.Sscanf(yn.Value, "%d", &v); err == nil {
			return v
		}
		return yn.Value
	case "!!float":
		var v float64
		if _, err := fmt.Sscanf(yn.Value, "%f", &v); err == nil {
			return v
		}
		return yn.Value
	default:
		// Try to infer type
		if yn.Value == "~" || yn.Value == "null" || yn.Value == "Null" || yn.Value == "NULL" {
			return nil
		}
		if yn.Value == "true" || yn.Value == "True" || yn.Value == "TRUE" {
			return true
		}
		if yn.Value == "false" || yn.Value == "False" || yn.Value == "FALSE" {
			return false
		}
		var intVal int
		if _, err := fmt.Sscanf(yn.Value, "%d", &intVal); err == nil {
			return intVal
		}
		var floatVal float64
		if _, err := fmt.Sscanf(yn.Value, "%f", &floatVal); err == nil {
			return floatVal
		}
		return yn.Value
	}
}

// CountDocuments counts the number of YAML documents in a file.
func CountDocuments(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	count := 0
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var doc yaml.Node
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}
