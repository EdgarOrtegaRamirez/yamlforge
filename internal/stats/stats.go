// Package stats provides YAML document statistics.
package stats

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"gopkg.in/yaml.v3"
)

// Engine computes statistics for YAML documents.
type Engine struct{}

// NewEngine creates a new stats engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Analyze analyzes YAML data and returns statistics.
func (e *Engine) Analyze(data []byte) (*models.YAMLStats, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	stats := &models.YAMLStats{
		KeyFrequency: make(map[string]int),
	}

	lines := strings.Split(string(data), "\n")
	stats.LineCount = len(lines)

	e.countNodes(&doc, stats, 0)
	return stats, nil
}

// AnalyzeFile analyzes a YAML file.
func (e *Engine) AnalyzeFile(path string) (*models.YAMLStats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return e.Analyze(data)
}

// countNodes recursively counts nodes in a YAML tree.
func (e *Engine) countNodes(node *yaml.Node, stats *models.YAMLStats, depth int) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		stats.TotalNodes++
		for _, child := range node.Content {
			e.countNodes(child, stats, depth)
		}
		return
	}

	stats.TotalNodes++

	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	switch node.Kind {
	case yaml.MappingNode:
		stats.MappingNodes++
		for i := 0; i < len(node.Content)-1; i += 2 {
			key := node.Content[i]
			val := node.Content[i+1]

			stats.TotalKeys++
			stats.KeyFrequency[key.Value]++

			e.countNodes(val, stats, depth+1)
		}

	case yaml.SequenceNode:
		stats.SequenceNodes++
		for _, child := range node.Content {
			e.countNodes(child, stats, depth+1)
		}

	case yaml.ScalarNode:
		stats.ScalarNodes++

	case yaml.AliasNode:
		stats.AliasCount++
		if node.Alias != nil {
			e.countNodes(node.Alias, stats, depth+1)
		}
	}

	if node.Anchor != "" {
		stats.AnchorCount++
	}
	if node.Tag != "" {
		stats.TagCount++
	}
}

// TopKeys returns the most frequently used keys.
func TopKeys(s *models.YAMLStats, n int) []KeyFreq {
	freqs := make([]KeyFreq, 0, len(s.KeyFrequency))
	for key, count := range s.KeyFrequency {
		freqs = append(freqs, KeyFreq{Key: key, Count: count})
	}

	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].Count > freqs[j].Count
	})

	if n > len(freqs) {
		n = len(freqs)
	}
	return freqs[:n]
}

// KeyFreq represents a key and its frequency.
type KeyFreq struct {
	Key   string
	Count int
}

// StatsString returns a human-readable string representation.
func StatsString(s *models.YAMLStats) string {
	var sb strings.Builder
	sb.WriteString("YAML Statistics\n")
	sb.WriteString("===============\n")
	sb.WriteString(fmt.Sprintf("Total nodes:     %d\n", s.TotalNodes))
	sb.WriteString(fmt.Sprintf("Mapping nodes:   %d\n", s.MappingNodes))
	sb.WriteString(fmt.Sprintf("Sequence nodes:  %d\n", s.SequenceNodes))
	sb.WriteString(fmt.Sprintf("Scalar nodes:    %d\n", s.ScalarNodes))
	sb.WriteString(fmt.Sprintf("Max depth:       %d\n", s.MaxDepth))
	sb.WriteString(fmt.Sprintf("Total keys:      %d\n", s.TotalKeys))
	sb.WriteString(fmt.Sprintf("Unique keys:     %d\n", len(s.KeyFrequency)))
	sb.WriteString(fmt.Sprintf("Line count:      %d\n", s.LineCount))
	sb.WriteString(fmt.Sprintf("Anchor count:    %d\n", s.AnchorCount))
	sb.WriteString(fmt.Sprintf("Alias count:     %d\n", s.AliasCount))
	sb.WriteString(fmt.Sprintf("Tag count:       %d\n", s.TagCount))

	if len(s.KeyFrequency) > 0 {
		sb.WriteString("\nTop Keys:\n")
		for _, kf := range TopKeys(s, 10) {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", kf.Key, kf.Count))
		}
	}

	return sb.String()
}

// StatsToMap converts stats to a map for JSON output.
func StatsToMap(s *models.YAMLStats) map[string]interface{} {
	return map[string]interface{}{
		"total_nodes":    s.TotalNodes,
		"mapping_nodes":  s.MappingNodes,
		"sequence_nodes": s.SequenceNodes,
		"scalar_nodes":   s.ScalarNodes,
		"max_depth":      s.MaxDepth,
		"total_keys":     s.TotalKeys,
		"unique_keys":    len(s.KeyFrequency),
		"line_count":     s.LineCount,
		"anchor_count":   s.AnchorCount,
		"alias_count":    s.AliasCount,
		"tag_count":      s.TagCount,
		"key_frequency":  s.KeyFrequency,
	}
}
