package convert

import (
	"encoding/json"
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/convert"
)

func TestConvert_ToJSON(t *testing.T) {
	yamlData := []byte(`name: test
version: 1.0
`)
	engine := convert.NewEngine()
	result, err := engine.ToJSON(yamlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if m["name"] != "test" {
		t.Fatalf("expected name='test', got %v", m["name"])
	}
}

func TestConvert_ToJSONCompact(t *testing.T) {
	yamlData := []byte(`name: test
version: 1.0
`)
	engine := convert.NewEngine()
	result, err := engine.ToJSONCompact(yamlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
}

func TestConvert_FromJSON(t *testing.T) {
	jsonData := []byte(`{"name": "test", "version": 1}`)
	engine := convert.NewEngine()
	result, err := engine.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestConvert_ToTOML(t *testing.T) {
	yamlData := []byte(`name: test
version: 1.0
`)
	engine := convert.NewEngine()
	result, err := engine.ToTOML(yamlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestConvert_ToCSV(t *testing.T) {
	yamlData := []byte(`- name: alice
  age: 30
- name: bob
  age: 25
`)
	engine := convert.NewEngine()
	result, err := engine.ToCSV(yamlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(result)
	if len(s) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestConvert_InvalidYAML(t *testing.T) {
	engine := convert.NewEngine()
	_, err := engine.ToJSON([]byte(`invalid: yaml: [unclosed`))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestConvert_InvalidJSON(t *testing.T) {
	engine := convert.NewEngine()
	_, err := engine.FromJSON([]byte(`{invalid json}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
