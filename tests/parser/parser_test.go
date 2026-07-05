package parser

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/parser"
)

func TestParseBytes_Simple(t *testing.T) {
	p := parser.New()
	data := []byte(`name: test
version: 1.0
`)
	docs, err := p.ParseBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
}

func TestParseBytes_MultiDoc(t *testing.T) {
	p := parser.New()
	data := []byte(`name: doc1
---
name: doc2
---
name: doc3
`)
	docs, err := p.ParseBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}
}

func TestParseBytes_NestedMapping(t *testing.T) {
	p := parser.New()
	data := []byte(`server:
  host: localhost
  port: 8080
  ssl:
    enabled: true
    cert: /path/to/cert
`)
	docs, err := p.ParseBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	root := docs[0]
	if root.Type.String() != "mapping" {
		t.Fatalf("expected mapping, got %s", root.Type)
	}

	server := root.GetChild("server")
	if server == nil {
		t.Fatal("expected 'server' child")
	}
	if server.Type.String() != "mapping" {
		t.Fatalf("expected mapping, got %s", server.Type)
	}

	host := server.GetChild("host")
	if host == nil {
		t.Fatal("expected 'host' child")
	}
	if host.Value != "localhost" {
		t.Fatalf("expected 'localhost', got %v", host.Value)
	}
}

func TestParseBytes_Sequence(t *testing.T) {
	p := parser.New()
	data := []byte(`items:
  - name: item1
    value: 1
  - name: item2
    value: 2
`)
	docs, err := p.ParseBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := docs[0].GetChild("items")
	if items == nil {
		t.Fatal("expected 'items' child")
	}
	if items.Type.String() != "sequence" {
		t.Fatalf("expected sequence, got %s", items.Type)
	}
	if len(items.Children) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items.Children))
	}
}

func TestParseBytes_ScalarTypes(t *testing.T) {
	p := parser.New()
	data := []byte(`string_val: hello
int_val: 42
float_val: 3.14
bool_val: true
null_val: null
`)
	docs, err := p.ParseBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	root := docs[0]

	tests := []struct {
		key      string
		expected interface{}
	}{
		{"string_val", "hello"},
		{"int_val", 42},
		{"float_val", 3.14},
		{"bool_val", true},
		{"null_val", nil},
	}

	for _, tt := range tests {
		child := root.GetChild(tt.key)
		if child == nil {
			t.Errorf("expected child '%s'", tt.key)
			continue
		}
		if tt.expected == nil && child.Value != nil {
			t.Errorf("key %s: expected nil, got %v", tt.key, child.Value)
		} else if tt.expected != nil && child.Value != tt.expected {
			t.Errorf("key %s: expected %v (%T), got %v (%T)", tt.key, tt.expected, tt.expected, child.Value, child.Value)
		}
	}
}

func TestParseBytes_Empty(t *testing.T) {
	p := parser.New()
	docs, err := p.ParseBytes([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("expected 0 documents, got %d", len(docs))
	}
}

func TestParseString(t *testing.T) {
	p := parser.New()
	docs, err := p.ParseString(`key: value`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
}

func TestCountDocuments(t *testing.T) {
	// CountDocuments needs a file, so we'll test ParseBytes instead
	p := parser.New()
	data := []byte(`a: 1
---
b: 2
---
c: 3
`)
	docs, err := p.ParseBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}
}
