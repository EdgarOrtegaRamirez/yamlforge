package format

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/format"
)

func TestFormat_PrettyPrint(t *testing.T) {
	data := []byte(`name: test
version: 1.0
`)
	result, err := format.PrettyPrint(data, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestFormat_SortKeys(t *testing.T) {
	data := []byte(`zebra: 1
alpha: 2
middle: 3
`)
	result, err := format.SortKeys(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(result)
	// alpha should come before zebra
	alphaIdx := 0
	zebraIdx := 0
	for i, c := range s {
		if c == 'a' && alphaIdx == 0 {
			alphaIdx = i
		}
		if c == 'z' && zebraIdx == 0 {
			zebraIdx = i
		}
	}
	if alphaIdx > zebraIdx {
		t.Fatal("expected alpha before zebra after sorting")
	}
}

func TestFormat_Minify(t *testing.T) {
	data := []byte(`name:   test
  
version:   1.0
`)
	result, err := format.Minify(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestFormat_WithIndent(t *testing.T) {
	data := []byte(`name: test
`)
	f := format.NewFormatter(format.Options{Indent: 4})
	result, err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestFormat_InvalidYAML(t *testing.T) {
	data := []byte(`invalid: yaml: [unclosed`)
	_, err := format.PrettyPrint(data, 2)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
