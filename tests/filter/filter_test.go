package filter

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/filter"
)

func TestFilter_CountDocuments(t *testing.T) {
	data := []byte(`a: 1
---
b: 2
---
c: 3
`)
	engine := filter.NewEngine()
	count, err := engine.CountDocuments(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 documents, got %d", count)
	}
}

func TestFilter_SplitDocuments(t *testing.T) {
	data := []byte(`a: 1
---
b: 2
`)
	engine := filter.NewEngine()
	docs, err := engine.SplitDocuments(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}
}

func TestFilter_ByContent(t *testing.T) {
	data := []byte(`name: alice
---
name: bob
---
name: alice
`)
	engine := filter.NewEngine()
	result, err := engine.FilterByContent(data, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(result)
	if len(s) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestFilter_ByIndex(t *testing.T) {
	data := []byte(`a: 1
---
b: 2
---
c: 3
`)
	engine := filter.NewEngine()
	result, err := engine.FilterByIndex(data, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(result)
	if s == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestFilter_ByIndex_OutOfRange(t *testing.T) {
	data := []byte(`a: 1`)
	engine := filter.NewEngine()
	_, err := engine.FilterByIndex(data, 5)
	if err == nil {
		t.Fatal("expected error for out of range index")
	}
}

func TestFilter_ByRange(t *testing.T) {
	data := []byte(`a: 1
---
b: 2
---
c: 3
`)
	engine := filter.NewEngine()
	result, err := engine.FilterByRange(data, 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(result)
	if len(s) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestFilter_ByRange_Invalid(t *testing.T) {
	data := []byte(`a: 1
---
b: 2
`)
	engine := filter.NewEngine()
	_, err := engine.FilterByRange(data, 2, 0)
	if err == nil {
		t.Fatal("expected error for invalid range")
	}
}

func TestFilter_SingleDocument(t *testing.T) {
	data := []byte(`name: test`)
	engine := filter.NewEngine()
	count, err := engine.CountDocuments(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 document, got %d", count)
	}
}
