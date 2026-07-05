package diff

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/diff"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/models"
	"github.com/EdgarOrtegaRamirez/yamlforge/internal/parser"
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

func TestDiff_Identical(t *testing.T) {
	a := parseYAML(t, `name: test
version: 1`)
	b := parseYAML(t, `name: test
version: 1`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	if diff.HasChanges(entries) {
		t.Fatal("expected no changes for identical documents")
	}
}

func TestDiff_AddedKey(t *testing.T) {
	a := parseYAML(t, `name: test`)
	b := parseYAML(t, `name: test
version: 1`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	added, _, _ := diff.Summary(entries)
	if added != 1 {
		t.Fatalf("expected 1 addition, got %d", added)
	}
}

func TestDiff_RemovedKey(t *testing.T) {
	a := parseYAML(t, `name: test
version: 1`)
	b := parseYAML(t, `name: test`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	_, removed, _ := diff.Summary(entries)
	if removed != 1 {
		t.Fatalf("expected 1 removal, got %d", removed)
	}
}

func TestDiff_ModifiedValue(t *testing.T) {
	a := parseYAML(t, `name: old`)
	b := parseYAML(t, `name: new`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	_, _, modified := diff.Summary(entries)
	if modified != 1 {
		t.Fatalf("expected 1 modification, got %d", modified)
	}
}

func TestDiff_NestedChanges(t *testing.T) {
	a := parseYAML(t, `
server:
  host: localhost
  port: 8080
`)
	b := parseYAML(t, `
server:
  host: example.com
  port: 443
`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	_, _, modified := diff.Summary(entries)
	if modified != 2 {
		t.Fatalf("expected 2 modifications, got %d", modified)
	}
}

func TestDiff_SequenceChanges(t *testing.T) {
	a := parseYAML(t, `
items:
  - a
  - b
`)
	b := parseYAML(t, `
items:
  - a
  - b
  - c
`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	added, _, _ := diff.Summary(entries)
	if added != 1 {
		t.Fatalf("expected 1 addition, got %d", added)
	}
}

func TestDiff_TypeChange(t *testing.T) {
	a := parseYAML(t, `value: string`)
	b := parseYAML(t, `value: 42`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	_, _, modified := diff.Summary(entries)
	if modified != 1 {
		t.Fatalf("expected 1 modification for type change, got %d", modified)
	}
}

func TestDiff_FormatText(t *testing.T) {
	a := parseYAML(t, `name: old`)
	b := parseYAML(t, `name: new`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	result := diff.FormatDiff(entries, "text")
	if result == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestDiff_FormatCompact(t *testing.T) {
	a := parseYAML(t, `name: old`)
	b := parseYAML(t, `name: new`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	result := diff.FormatDiff(entries, "compact")
	if result == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestDiff_FormatJSON(t *testing.T) {
	a := parseYAML(t, `name: old`)
	b := parseYAML(t, `name: new`)
	engine := diff.NewEngine()
	entries := engine.Diff(a, b)
	result := diff.FormatDiff(entries, "json")
	if result == "" {
		t.Fatal("expected non-empty output")
	}
}
