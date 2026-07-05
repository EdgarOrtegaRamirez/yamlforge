package validate

import (
	"testing"

	"github.com/EdgarOrtegaRamirez/yamlforge/internal/validate"
)

func TestValidate_RequiredFields(t *testing.T) {
	schema := &validate.Schema{
		Required: []string{"name", "version"},
	}

	data := []byte(`name: test`)
	engine := validate.NewEngine()
	report, err := engine.Validate(data, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.HasErrors() {
		t.Fatal("expected validation error for missing required field")
	}
}

func TestValidate_TypeCheck(t *testing.T) {
	schema := &validate.Schema{
		Types: map[string]string{
			"name":    "string",
			"version": "integer",
		},
	}

	data := []byte(`name: test
version: not-a-number
`)
	engine := validate.NewEngine()
	report, err := engine.Validate(data, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "type-check" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected type-check error")
	}
}

func TestValidate_EnumCheck(t *testing.T) {
	schema := &validate.Schema{
		Enums: map[string][]string{
			"env": {"dev", "staging", "prod"},
		},
	}

	data := []byte(`env: invalid`)
	engine := validate.NewEngine()
	report, err := engine.Validate(data, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "enum-check" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected enum-check error")
	}
}

func TestValidate_MinLength(t *testing.T) {
	schema := &validate.Schema{
		MinLength: map[string]int{
			"name": 3,
		},
	}

	data := []byte(`name: ab`)
	engine := validate.NewEngine()
	report, err := engine.Validate(data, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "min-length" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected min-length error")
	}
}

func TestValidate_MaxLength(t *testing.T) {
	schema := &validate.Schema{
		MaxLength: map[string]int{
			"name": 3,
		},
	}

	data := []byte(`name: toolong`)
	engine := validate.NewEngine()
	report, err := engine.Validate(data, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, issue := range report.Issues {
		if issue.Rule == "max-length" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected max-length error")
	}
}

func TestValidate_Passing(t *testing.T) {
	schema := &validate.Schema{
		Required: []string{"name"},
		Types: map[string]string{
			"name": "string",
		},
	}

	data := []byte(`name: test`)
	engine := validate.NewEngine()
	report, err := engine.Validate(data, schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.HasErrors() {
		t.Fatalf("expected no errors, got %d", report.ErrorCount())
	}
}

func TestInferSchema(t *testing.T) {
	data := []byte(`name: test
version: 1.0
enabled: true
`)
	schema, err := validate.InferSchema(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schema.Required) != 3 {
		t.Fatalf("expected 3 required fields, got %d", len(schema.Required))
	}

	if schema.Types["name"] != "string" {
		t.Fatalf("expected name type 'string', got %s", schema.Types["name"])
	}
}
