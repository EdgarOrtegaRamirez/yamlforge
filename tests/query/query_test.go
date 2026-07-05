package query

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/parser"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/query"
)

func parseYAML(t *testing.T, data string) *models.Node {
	t.Helper()
	p := parser.New()
	docs, err := p.ParseString(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("no documents parsed")
	}
	return docs[0]
}

func TestQuery_SimpleKey(t *testing.T) {
	root := parseYAML(t, `name: test`)
	engine := query.NewEngine()
	val, err := engine.QueryValue(root, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "test" {
		t.Fatalf("expected 'test', got %v", val)
	}
}

func TestQuery_NestedKey(t *testing.T) {
	root := parseYAML(t, `
server:
  host: localhost
  port: 8080
`)
	engine := query.NewEngine()
	val, err := engine.QueryValue(root, "server.host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "localhost" {
		t.Fatalf("expected 'localhost', got %v", val)
	}
}

func TestQuery_ArrayIndex(t *testing.T) {
	root := parseYAML(t, `
items:
  - apple
  - banana
  - cherry
`)
	engine := query.NewEngine()
	val, err := engine.QueryValue(root, "items[1]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "banana" {
		t.Fatalf("expected 'banana', got %v", val)
	}
}

func TestQuery_Wildcard(t *testing.T) {
	root := parseYAML(t, `
a: 1
b: 2
c: 3
`)
	engine := query.NewEngine()
	results, err := engine.Query(root, "*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestQuery_NotFound(t *testing.T) {
	root := parseYAML(t, `name: test`)
	engine := query.NewEngine()
	_, err := engine.QueryValue(root, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}

func TestQuery_SetValue(t *testing.T) {
	root := parseYAML(t, `name: old`)
	engine := query.NewEngine()
	err := engine.SetValue(root, "name", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := engine.QueryValue(root, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "new" {
		t.Fatalf("expected 'new', got %v", val)
	}
}

func TestParsePath_Simple(t *testing.T) {
	segments, err := query.ParsePath("name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	if segments[0].Name != "name" {
		t.Fatalf("expected 'name', got %s", segments[0].Name)
	}
}

func TestParsePath_Dotted(t *testing.T) {
	segments, err := query.ParsePath("server.host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}
}

func TestParsePath_Index(t *testing.T) {
	segments, err := query.ParsePath("items[0]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}
	if segments[0].Name != "items" {
		t.Fatalf("expected first segment 'items', got %s", segments[0].Name)
	}
	if !segments[1].IsIndex {
		t.Fatal("expected second segment to be index")
	}
	if segments[1].Index != 0 {
		t.Fatalf("expected index 0, got %d", segments[1].Index)
	}
}

func TestQuery_DeepNesting(t *testing.T) {
	root := parseYAML(t, `
level1:
  level2:
    level3:
      value: deep
`)
	engine := query.NewEngine()
	val, err := engine.QueryValue(root, "level1.level2.level3.value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "deep" {
		t.Fatalf("expected 'deep', got %v", val)
	}
}
