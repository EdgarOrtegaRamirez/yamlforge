package merge

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/merge"
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

func TestMerge_Deep(t *testing.T) {
	a := parseYAML(t, `
server:
  host: localhost
  port: 8080
`)
	b := parseYAML(t, `
server:
  host: example.com
  ssl: true
`)
	engine := merge.NewEngine(models.MergeDeep)
	result, err := engine.Merge(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	server := result.GetChild("server")
	if server == nil {
		t.Fatal("expected 'server' child")
	}

	host := server.GetChild("host")
	if host == nil || host.Value != "example.com" {
		t.Fatalf("expected host to be 'example.com', got %v", host.Value)
	}

	port := server.GetChild("port")
	if port == nil || port.Value != 8080 {
		t.Fatalf("expected port to be 8080, got %v", port.Value)
	}

	ssl := server.GetChild("ssl")
	if ssl == nil || ssl.Value != true {
		t.Fatalf("expected ssl to be true, got %v", ssl.Value)
	}
}

func TestMerge_Replace(t *testing.T) {
	a := parseYAML(t, `name: old`)
	b := parseYAML(t, `name: new`)
	engine := merge.NewEngine(models.MergeReplace)
	result, err := engine.Merge(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name := result.GetChild("name")
	if name == nil || name.Value != "new" {
		t.Fatalf("expected 'new', got %v", name.Value)
	}
}

func TestMerge_Keep(t *testing.T) {
	a := parseYAML(t, `name: original`)
	b := parseYAML(t, `name: new
version: 1`)
	engine := merge.NewEngine(models.MergeKeep)
	result, err := engine.Merge(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name := result.GetChild("name")
	if name == nil || name.Value != "original" {
		t.Fatalf("expected 'original', got %v", name.Value)
	}
	version := result.GetChild("version")
	if version == nil || version.Value != 1 {
		t.Fatalf("expected version 1, got %v", version.Value)
	}
}

func TestMerge_Append(t *testing.T) {
	a := parseYAML(t, `
items:
  - a
`)
	b := parseYAML(t, `
items:
  - b
`)
	engine := merge.NewEngine(models.MergeAppend)
	result, err := engine.Merge(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := result.GetChild("items")
	if items == nil {
		t.Fatal("expected 'items' child")
	}
	if len(items.Children) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items.Children))
	}
}

func TestMerge_NilSource(t *testing.T) {
	a := parseYAML(t, `name: test`)
	engine := merge.NewEngine(models.MergeDeep)
	result, err := engine.Merge(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name := result.GetChild("name")
	if name == nil || name.Value != "test" {
		t.Fatalf("expected 'test', got %v", name.Value)
	}
}

func TestMerge_NilTarget(t *testing.T) {
	b := parseYAML(t, `name: test`)
	engine := merge.NewEngine(models.MergeDeep)
	result, err := engine.Merge(nil, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name := result.GetChild("name")
	if name == nil || name.Value != "test" {
		t.Fatalf("expected 'test', got %v", name.Value)
	}
}
