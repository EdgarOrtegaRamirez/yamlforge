package stats

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/stats"
)

func TestStats_Analyze(t *testing.T) {
	data := []byte(`server:
  host: localhost
  port: 8080
  ssl:
    enabled: true
items:
  - name: item1
  - name: item2
`)
	engine := stats.NewEngine()
	s, err := engine.Analyze(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.MappingNodes == 0 {
		t.Fatal("expected at least 1 mapping node")
	}
	if s.SequenceNodes == 0 {
		t.Fatal("expected at least 1 sequence node")
	}
	if s.ScalarNodes == 0 {
		t.Fatal("expected at least 1 scalar node")
	}
	if s.TotalKeys == 0 {
		t.Fatal("expected at least 1 key")
	}
	if s.MaxDepth == 0 {
		t.Fatal("expected depth > 0")
	}
	if s.LineCount == 0 {
		t.Fatal("expected line count > 0")
	}
}

func TestStats_TopKeys(t *testing.T) {
	data := []byte(`a: 1
b: 2
a: 3
c: 4
`)
	engine := stats.NewEngine()
	s, err := engine.Analyze(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	topKeys := stats.TopKeys(s, 3)
	if len(topKeys) == 0 {
		t.Fatal("expected at least 1 top key")
	}
}

func TestStats_StatsString(t *testing.T) {
	data := []byte(`name: test`)
	engine := stats.NewEngine()
	s, err := engine.Analyze(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := stats.StatsString(s)
	if len(result) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestStats_StatsToMap(t *testing.T) {
	data := []byte(`name: test`)
	engine := stats.NewEngine()
	s, err := engine.Analyze(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := stats.StatsToMap(s)
	if m == nil {
		t.Fatal("expected non-nil map")
	}
}

func TestStats_Empty(t *testing.T) {
	data := []byte(``)
	engine := stats.NewEngine()
	s, err := engine.Analyze(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.MappingNodes != 0 {
		t.Fatalf("expected 0 mapping nodes for empty YAML, got %d", s.MappingNodes)
	}
	if s.SequenceNodes != 0 {
		t.Fatalf("expected 0 sequence nodes for empty YAML, got %d", s.SequenceNodes)
	}
	if s.ScalarNodes != 0 {
		t.Fatalf("expected 0 scalar nodes for empty YAML, got %d", s.ScalarNodes)
	}
}

func TestStats_Nested(t *testing.T) {
	data := []byte(`a:
  b:
    c: 1
`)
	engine := stats.NewEngine()
	s, err := engine.Analyze(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.MaxDepth < 3 {
		t.Fatalf("expected depth >= 3, got %d", s.MaxDepth)
	}
}
